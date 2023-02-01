package powproto

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
	QuoteRequest
	QuoteResponse
	ExitRequest
)

const MessageTerminator byte = '\n'

var (
	ErrZeroLengthData     = errors.New("powproto: zero length data")
	ErrInvalidMessageKind = errors.New("powproto: invalid message kind")
)

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
		return nil, ErrZeroLengthData
	}

	kind, err := strconv.Atoi(string(data[0]))
	if err != nil {
		return nil, ErrInvalidMessageKind
	}

	msg := Message{
		Kind: MessageKind(kind),
	}
	if len(data) > 1 {
		msg.Payload = data[1:]
	}
	return &msg, nil
}
