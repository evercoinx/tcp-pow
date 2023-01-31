package proto

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MessageKind int

const (
	ChallengeRequest MessageKind = iota
	ChallengeResponse
	ResourceRequest
	ResourceResponse
	ExitRequest
)

const MessageTerminator byte = '\n'

type Message struct {
	Kind    MessageKind
	Payload string
}

func NewMessage(kind MessageKind, payload string) Message {
	return Message{
		Kind:    kind,
		Payload: payload,
	}
}

func (m Message) String() string {
	return fmt.Sprintf("%d%s", m.Kind, m.Payload)
}

func Parse(data string) (*Message, error) {
	data = strings.TrimRight(data, string(MessageTerminator))
	if len(data) == 0 {
		return nil, errors.New("zero length message")
	}

	kind, err := strconv.Atoi(string(data[0]))
	if err != nil {
		return nil, errors.New("invalid message kind")
	}

	msg := Message{
		Kind: MessageKind(kind),
	}
	if len(data) > 1 {
		msg.Payload = data[1:]
	}
	return &msg, nil
}
