package tcpclient

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/proto"
	log "github.com/sirupsen/logrus"
)

func Query(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := handleConnection(conn); err != nil {
		return err
	}
	return nil
}

func handleConnection(conn net.Conn) error {
	challengeReqMsg := proto.NewMessage(proto.ChallengeRequest, "")
	if err := writeMessage(challengeReqMsg, conn); err != nil {
		return fmt.Errorf("failed to write challenge message: %w", err)
	}

	r := bufio.NewReader(conn)
	challengeResData, err := r.ReadString(proto.MessageTerminator)
	if err != nil {
		return fmt.Errorf("failed to read challenge data: %w", err)
	}

	msg, err := proto.Parse(challengeResData)
	if err != nil {
		return fmt.Errorf("failed to parse challenge data : %w", err)
	}

	hc, err := hashcash.Parse(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse hashcash: %w", err)
	}
	log.WithField("hashcash", hc).Debug("parsed hashcash")

	exitReqMsg := proto.NewMessage(proto.ExitRequest, "")
	if err := writeMessage(exitReqMsg, conn); err != nil {
		return fmt.Errorf("failed to write exit message: %w", err)
	}

	log.Debug("disconnected")
	return nil
}

func writeMessage(msg proto.Message, conn io.Writer) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), proto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
