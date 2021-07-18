package stream

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var (
	// ErrRevN is an error that is returned when stream is unable to
	// read stream in reversed order.
	ErrRevN = errors.New("unable to read reversed stream")
	// ErrPublish is an error that is returned when stream is unable
	// to put message to underlying data storage.
	ErrPublish = errors.New("unable to publish message")
)

// StreamOptions provides options to configure Stream.
type StreamOptions struct {
	// OperationTimeout is the maximum number of time is allowed to wait
	// until stream operation is completed. When operation timeout is reached,
	// underlying context is canceled and operation returns an error.
	OperationTimeout time.Duration
	// StreamMaxLength sets the MAXLEN option when calling XADD. This creates a
	// capped stream to prevent the stream from taking up memory indefinitely.
	StreamMaxLength int64
	// Approx determines whether to use the ~ with the MAXLEN
	// option. This allows the stream trimming to done in a more efficient
	// manner.
	Approx bool
	// RedisClient allows you to inject an already-made Redis Client for use in
	// the consumer.
	RedisClient redis.UniversalClient
	// RedisOptions allows you to configure the underlying Redis connection.
	RedisOptions *RedisOptions
}

var defaultStreamOptions = &StreamOptions{
	StreamMaxLength:  100,
	OperationTimeout: 5 * time.Second,
}

// Stream provides convenient functions to interact with
// underlying stream data storage.
type Stream interface {
	RevN(ch chan<- []byte, n int64) error
	Publish(msg []byte) error
}

// stream represents named Redis stream which implements
// Stream interface.
type stream struct {
	name string

	redis   redis.UniversalClient
	options *StreamOptions
}

// NewStream creates new stream with default options.
func NewStream(name string) Stream {
	return NewStreamWithOptions(name, defaultStreamOptions)
}

// NewStreamWithOptions creates new stream with provided options.
// When OperationTimeout or StreamMaxLength is set to zero,
// default options is set.
func NewStreamWithOptions(name string, options *StreamOptions) Stream {
	var r redis.UniversalClient
	if options.RedisClient != nil {
		r = options.RedisClient
	} else {
		r = newRedisClient(options.RedisOptions)
	}
	if options.StreamMaxLength == 0 {
		options.StreamMaxLength = defaultStreamOptions.StreamMaxLength
	}
	if options.OperationTimeout == 0 {
		options.OperationTimeout = defaultStreamOptions.OperationTimeout
	}
	return &stream{
		name:    name,
		options: options,
		redis:   r,
	}
}

// RevN writes partial reversed stream to the channel.
func (s *stream) RevN(ch chan<- []byte, n int64) error {
	ctx, cancel := s.newTimeoutContext()
	defer cancel()

	// Operation will timeout in several seconds depending on
	// stream configuration.
	res, err := s.redis.XRevRangeN(ctx, s.name, "+", "-", n).Result()
	if err != nil {
		logrus.Error(err)
		return ErrRevN
	}
	for i := len(res) - 1; i >= 0; i-- {
		if p, ok := res[i].Values["message"]; ok {
			message, ok := p.(string)
			if !ok {
				continue
			}
			ch <- []byte(message)
		}
	}
	return nil
}

// Publish the message to the stream.
func (s *stream) Publish(msg []byte) error {
	ctx, cancel := s.newTimeoutContext()
	defer cancel()

	args := &redis.XAddArgs{
		Stream: s.name,
		Values: map[string]interface{}{
			"message": string(msg),
		},
	}
	// Set default options if applicable.
	if s.options.Approx {
		args.Approx = s.options.Approx
	} else {
		args.MaxLen = s.options.StreamMaxLength
	}
	// Add the message to the stream.
	_, err := s.redis.XAdd(ctx, args).Result()
	if err != nil {
		logrus.Error(err)
		return ErrPublish
	}
	return nil
}

// newTimeoutContext creates new context with default stream operation
// timeout.
func (s *stream) newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.options.OperationTimeout)
}
