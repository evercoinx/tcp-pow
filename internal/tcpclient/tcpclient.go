package tcpclient

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/powproto"
	log "github.com/sirupsen/logrus"
)

func QueryPipeline(address string) error {
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

	if _, err := requestQuote(conn, hc); err != nil {
		return err
	}

	if err := requestExit(conn); err != nil {
		return err
	}
	return nil
}

func requestChallenge(conn net.Conn) (*hashcash.Hashcash, error) {
	reqMsg := powproto.NewMessage(powproto.ChallengeRequest, "")
	if err := writeMessage(reqMsg, conn); err != nil {
		return nil, fmt.Errorf("failed to write challenge request message: %w", err)
	}
	log.Info("challenge requested")

	r := bufio.NewReader(conn)
	resData, err := r.ReadString(powproto.MessageTerminator)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge response data: %w", err)
	}

	resMsg, err := powproto.Parse(resData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse challenge response data : %w", err)
	}

	hc, err := hashcash.Unmarshal(resMsg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hashcash payload: %w", err)
	}
	log.WithField("initial_challenge", hc).Info("challenge responded")

	return hc, err
}

func requestQuote(conn net.Conn, hc *hashcash.Hashcash) (string, error) {
	if err := hc.Compute(); err != nil {
		return "", fmt.Errorf("failed to compute challenge: %w", err)
	}

	reqMsg := powproto.NewMessage(powproto.QuoteRequest, hc.String())
	if err := writeMessage(reqMsg, conn); err != nil {
		return "", fmt.Errorf("failed to write quote request message: %w", err)
	}
	log.WithField("computed_challenge", hc.String()).Info("quote requested")

	r := bufio.NewReader(conn)
	resData, err := r.ReadString(powproto.MessageTerminator)
	if err != nil {
		return "", fmt.Errorf("failed to read quote response data: %w", err)
	}

	resMsg, err := powproto.Parse(resData)
	if err != nil {
		return "", fmt.Errorf("failed to parse quote response data : %w", err)
	}
	log.WithFields(log.Fields{"quote": resMsg.Payload}).Info("quote responded")
	return resMsg.Payload, nil
}

func requestExit(conn net.Conn) error {
	reqMsg := powproto.NewMessage(powproto.ExitRequest, "")
	if err := writeMessage(reqMsg, conn); err != nil {
		return fmt.Errorf("failed to write exit request message: %w", err)
	}
	log.Info("exit requested")
	return nil
}

func writeMessage(msg powproto.Message, conn io.Writer) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), powproto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
