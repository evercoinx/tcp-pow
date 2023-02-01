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
	hc, err := requestChallenge(conn)
	if err != nil {
		return err
	}
	log.WithField("challenge", hc).Debug("challenge requested")

	rs, err := requestResource(conn, hc)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"iterations": hc.Counter, "resource": rs}).Debug("resource requested")

	if err := requestExit(conn); err != nil {
		return err
	}
	log.Debug("exit requested")
	return nil
}

func requestChallenge(conn net.Conn) (*hashcash.Hashcash, error) {
	reqMsg := proto.NewMessage(proto.ChallengeRequest, "")
	if err := writeMessage(reqMsg, conn); err != nil {
		return nil, fmt.Errorf("failed to write challenge request message: %w", err)
	}

	r := bufio.NewReader(conn)
	resData, err := r.ReadString(proto.MessageTerminator)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge response data: %w", err)
	}

	resMsg, err := proto.Parse(resData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse challenge response data : %w", err)
	}

	challenge, err := hashcash.Parse(resMsg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hashcash string: %w", err)
	}
	return challenge, err
}

func requestResource(conn net.Conn, hc *hashcash.Hashcash) (string, error) {
	challenge, err := hc.Compute()
	if err != nil {
		return "", fmt.Errorf("failed to compute hashcash: %w", err)
	}

	reqMsg := proto.NewMessage(proto.ResourceRequest, challenge)
	if err := writeMessage(reqMsg, conn); err != nil {
		return "", fmt.Errorf("failed to write resource request message: %w", err)
	}
	return "", nil
}

func requestExit(conn net.Conn) error {
	reqMsg := proto.NewMessage(proto.ExitRequest, "")
	if err := writeMessage(reqMsg, conn); err != nil {
		return fmt.Errorf("failed to write exit request message: %w", err)
	}
	return nil
}

func writeMessage(msg proto.Message, conn io.Writer) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), proto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
