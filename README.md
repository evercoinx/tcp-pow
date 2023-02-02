# Task Description
Design and implement “Word of Wisdom” TCP server.
- TCP server should be protected from DDOS attacks with the Prof of Work, the challenge-response protocol should be used.
- The choice of the POW algorithm should be explained.
- After Prof Of Work verification, the server should send one of the quotes from “Word of wisdom” book or any other collection of the quotes.
- Docker file should be provided both for the server and for the client that solves the PoW challenge

# Implementation Description
The client and server communicate with messages over a application-level challenge-response POW protocol running over the TCP/IP stack. The message consists of two parts -- the header and the payload. To prevent DDOS attacks the POW protocol relies on the [Hashcash mechanism](https://en.wikipedia.org/wiki/Hashcash), which additionally allows to control client's challenge complexity and imposes client's request expiration.

## Message Header
The message header is required. It contains a single field `MessageKind`. The length of the header is 1-byte. It allows to discriminate message types during communication between parties. A list of possible `MessageKinds`:
- ChallengeRequest = 0
- ChallengeResponse = 1
- QuoteRequest = 2
- QuoteResponse = 3
- ExitRequest = 4

## Message Payload
The message payload is optional. It may have an arbitrary length and is subject to the upper bound restriction of 4Kb.

# Client-Server Communication
1. The client sends the `ChallageRequest` message to the server.
2. The server generates the `ChallengeResponse` message containing a PoW challenge, which is a Hashcash string.
3. The client solves the PoW challenge and sends its proof in the `QuoteRequest` message to the server.
4. The server verifies the solved PoW challenge and if it is a success sends one of the quotes in the `QuoteResponse` message to the client.
5. The client may notify the server about disconnection with the `ExitRequest` message.

# How to Run
The command starts the TCP server and the TCP client. The client periodically pings the server with a request pipeline. 
```
make service-up
```

# How to Stop
```
make service-down
```
