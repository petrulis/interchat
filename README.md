# Interhat

Chat service with configurable message history.

## Design Decisions

1. Service opens a pool of Redis connections for a room and supports reading the stream and writing to it
   using stream.Stream abstraction.
2. Since there is no need to support neither multiple channels nor authenticated users, message structure
   kept as simple as possible.
