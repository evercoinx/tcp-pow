package tcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/proto"
	log "github.com/sirupsen/logrus"
)

func Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start tcp listener: %w", err)
	}
	defer l.Close()

	log.WithField("address", l.Addr).Info("server listening")
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("failed to create new tcp connection: %w", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr()
	log.WithField("address", addr).Debug("client connected")

	r := bufio.NewReader(conn)
	for {
		reqData, err := r.ReadString(proto.MessageTerminator)
		if err != nil {
			log.WithError(err).Error("failed to read request data")
			return
		}

		msg, err := constructMessage(reqData, addr)
		if err != nil {
			log.WithError(err).Error("failed to construct message")
			return
		}
		if msg == nil {
			log.WithField("address", addr).Debug("client disconnected")
			return
		}

		if err := writeMessage(*msg, conn); err != nil {
			log.WithError(err).Error("failed to write response message")
			return
		}
	}
}

func constructMessage(requestData string, clientAddr net.Addr) (*proto.Message, error) {
	msg, err := proto.Parse(requestData)
	if err != nil {
		return nil, err
	}

	switch msg.Kind {
	case proto.ChallengeRequest:
		resource := strings.Replace(clientAddr.String(), ":", "/", 1)
		hc, err := hashcash.NewHashcash(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to generate hashcash: %w", err)
		}

		msg := proto.Message{
			Kind:    proto.ChallengeResponse,
			Payload: hc.String(),
		}
		return &msg, nil
	case proto.ExitRequest:
		return nil, nil
	default:
		return nil, errors.New("unknown message kind")
	}
}

func writeMessage(msg proto.Message, conn net.Conn) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), proto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
