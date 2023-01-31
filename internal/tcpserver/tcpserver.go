package tcpserver

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/proto"
)

func Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer l.Close()

	fmt.Printf("server listening at %s\n", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr()
	fmt.Printf("client %s connected\n", addr)

	r := bufio.NewReader(conn)
	for {
		reqData, err := r.ReadString(proto.MessageTerminator)
		if err != nil {
			fmt.Println("failed to read request data:", err)
			return
		}

		msg, err := constructMessage(reqData, addr)
		if err != nil {
			fmt.Println("failed to construct message:", err)
			return
		}
		if msg == nil {
			fmt.Printf("client %s exited\n", addr)
			return
		}

		if err := writeMessage(*msg, conn); err != nil {
			fmt.Println("failed to write message:", err)
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
