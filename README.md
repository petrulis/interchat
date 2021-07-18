# Interchat

Chat service with configurable message history.

## Design Decisions

1. Since only one room is supported, protocol kept simple, meaning that connection upgrade stands
   for channel subscription and error on websocket connection stands for leaving the chat.
2. Service opens a pool of Redis connections for a room and supports reading the stream and writing to it
   using stream.Stream abstraction.
3. Since there is no need to support neither multiple channels nor authenticated users, message structure
   kept as simple as possible.
4. Message history is written to Redis and read with XREVRANGEN command.
5. Messages are added to history with XADD.
