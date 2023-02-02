# Task Description
Design and implement “Word of Wisdom” TCP server.
- TCP server should be protected from DDOS attacks with the Prof of Work, the challenge-response protocol should be used.
- The choice of the POW algorithm should be explained.
- After Prof Of Work verification, the server should send one of the quotes from “Word of wisdom” book or any other collection of the quotes.
- Docker file should be provided both for the server and for the client that solves the PoW challenge

# Implementation Description
The client and server communicate with messages over a application-level challenge-response POW protocol running over the TCP/IP stack. The message consists of two parts -- the header and the payload. To prevent DDOS attacks the POW protocol relies on the [Hashcash mechanism](https://en.wikipedia.org/wiki/Hashcash), which allows to control client's challenge complexity and imposes client's request expiration.

## Message header
The message header contains a single field `MessageKind`. The length of the header is 1-byte. It allows to discriminate message types during server-client interaction. Possible `MessageKinds` are:
- ChallengeRequest = 0
- ChallengeResponse = 1
- QuoteRequest = 2
- QuoteResponse = 3
- ExitRequest = 4

## Message payload
The message payload is allowed to have an arbitrary length and is subject to the upper bound restriction of 4Kb.

# Client-Server Interaction
1. The client sends the `ChallageRequest` message to server.
2. The server generates the `ChallengeResponse` message containing a PoW challenge, whtich is a Hashcash string.
3. The client solves the PoW challenge and sends the proof in the `QuoteRequest` message.
4. The server verifies the solved PoW challenge and in case of a success sends one of quotes of wisdom to the client in the `QuoteResponse` message.
5. The client notifies the server about disconnection with the `ExitRequest` message.

# How to Run
The command starts the TCP server and the TCP client which periodically pings the server with a request pipeline. 
```
make service-up
```

# How to Stop
```
make service-down
```
