package stream

import (
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestNewStream(t *testing.T) {
	t.Run("creates new client with default options", func(tt *testing.T) {
		s := NewStream("name")
		assert.NotNil(tt, s)
	})
}

func TestMultipleRevN(t *testing.T) {
	type testcase struct {
		num      int
		expected int
	}
	var cases []testcase
	for n := 0; n < 100; n++ {
		cases = append(cases, testcase{num: n, expected: n})
	}
	for _, tc := range cases {
		db, mock := redismock.NewClientMock()

		var result []redis.XMessage
		for n := 0; n < tc.num; n++ {
			result = append(result, redis.XMessage{
				Values: map[string]interface{}{
					"message": "",
				},
			})
		}

		mock.ExpectXRevRangeN("name", "+", "-", int64(tc.num)).SetVal(result)

		s := NewStreamWithOptions("name", &StreamOptions{
			RedisClient: db,
		})
		ch := make(chan []byte)
		go func() {
			defer close(ch)
			s.RevN(ch, int64(tc.num))
		}()
		received := 0
		for range ch {
			received++
		}
		assert.Equal(t, tc.expected, received)
	}
}

func TestRevNWithError(t *testing.T) {
	db, mock := redismock.NewClientMock()

	mock.ExpectXRevRangeN("name", "+", "-", 0).SetErr(fmt.Errorf(""))

	s := NewStreamWithOptions("name", &StreamOptions{
		RedisClient: db,
	})
	ch := make(chan []byte)
	err := s.RevN(ch, 0)
	assert.Equal(t, ErrRevN, err)
}

func TestPublish(t *testing.T) {
	t.Run("publishes message with no error", func(tt *testing.T) {
		db, mock := redismock.NewClientMock()

		args := &redis.XAddArgs{
			Stream: "name",
			Values: map[string]interface{}{
				"message": "",
			},
			Approx: false,
			MaxLen: defaultStreamOptions.StreamMaxLength,
		}
		mock.ExpectXAdd(args).SetVal("0")

		s := NewStreamWithOptions("name", &StreamOptions{
			RedisClient: db,
		})
		err := s.Publish([]byte{})
		assert.Equal(tt, nil, err)
	})

	t.Run("publishes message with error", func(tt *testing.T) {
		db, mock := redismock.NewClientMock()

		args := &redis.XAddArgs{}
		mock.ExpectXAdd(args).SetErr(fmt.Errorf(""))

		s := NewStreamWithOptions("name", &StreamOptions{
			RedisClient: db,
		})
		err := s.Publish([]byte{})
		assert.Equal(tt, ErrPublish, err)
	})
}
