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

	log.WithField("address", address).Info("server listening")
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
	reqMsg, err := proto.Parse(requestData)
	if err != nil {
		return nil, err
	}

	switch reqMsg.Kind {
	case proto.ChallengeRequest:
		resMsg, err := constructChallengeResponse(clientAddr)
		if err != nil {
			return nil, err
		}
		return resMsg, nil
	case proto.ResourceRequest:
		return nil, nil
	case proto.ExitRequest:
		return nil, nil
	case proto.ChallengeResponse, proto.ResourceResponse:
		return nil, errors.New("invalid message kind")
	}
	return nil, errors.New("unsupported message kind")
}

func constructChallengeResponse(clientAddr net.Addr) (*proto.Message, error) {
	resource := strings.Replace(clientAddr.String(), ":", "/", 1)
	hc, err := hashcash.NewHashcash(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hashcash: %w", err)
	}

	return &proto.Message{
		Kind:    proto.ChallengeResponse,
		Payload: hc.String(),
	}, nil
}

func writeMessage(msg proto.Message, conn net.Conn) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), proto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
