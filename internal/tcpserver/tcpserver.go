package tcpserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/powproto"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const cacheKeyPrefix = "tcp-pow-server:"

var (
	ErrUnexpectedMessageKind  = errors.New("unexpected message kind")
	ErrUnsupportedMessageKind = errors.New("unsupported message kind")
	ErrInvalidChallengeRand   = errors.New("invalid challenge rand")
)

var quotes = []string{
	"The journey of a thousand miles begins with one step.",
	"He is no fool who gives what he cannot keep to gain what he cannot lose.",
	"It's not what you look at that matters, it's what you see.",
	"A man should always consider how much he has more than he wants.",
	"That old law about 'an eye for an eye' leaves everybody blind. The time is always right to do the right thing.",
}

type Server struct {
	cacheClient     *redis.Client
	cacheExpiration time.Duration
}

func NewServer(c *redis.Client, cacheExpiration time.Duration) *Server {
	return &Server{
		cacheClient:     c,
		cacheExpiration: cacheExpiration,
	}
}

func (s *Server) Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start tcp listener: %w", err)
	}
	defer l.Close()

	log.WithField("server_address", address).Info("server listening")
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("failed to create new tcp connection: %w", err)
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr()
	log.WithField("client_address", addr).Info("client connected")

	r := bufio.NewReader(conn)
	for {
		reqData, err := r.ReadString(powproto.MessageTerminator)
		if err != nil {
			log.WithError(err).Error("failed to read request data")
			return
		}

		resMsg, err := s.processMessage(context.Background(), reqData, addr)
		if err != nil {
			log.WithError(err).Error("failed to process message")
			return
		}
		if resMsg == nil {
			log.WithField("client_address", addr).Info("client disconnected")
			return
		}

		if err := writeMessage(*resMsg, conn); err != nil {
			log.WithError(err).Error("failed to write response message")
			return
		}
	}
}

func (s *Server) processMessage(ctx context.Context, requestData string, clientAddr net.Addr) (*powproto.Message, error) {
	reqMsg, err := powproto.Parse(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request data: %w", err)
	}

	switch reqMsg.Kind {
	case powproto.ChallengeRequest:
		resMsg, err := s.processChallenge(ctx, clientAddr)
		if err != nil {
			return nil, err
		}
		return resMsg, nil
	case powproto.QuoteRequest:
		resMsg, err := s.processQuote(ctx, reqMsg.Payload, clientAddr)
		if err != nil {
			return nil, err
		}
		return resMsg, nil
	case powproto.ExitRequest:
		return nil, nil
	case powproto.ChallengeResponse, powproto.QuoteResponse:
		return nil, ErrUnexpectedMessageKind
	}
	return nil, ErrUnsupportedMessageKind
}

func (s *Server) processChallenge(ctx context.Context, clientAddr net.Addr) (*powproto.Message, error) {
	hc, err := hashcash.NewHashcash(clientAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, clientAddr.String())
	if err := s.cacheClient.SetEx(ctx, cacheKey, hc.Rand, s.cacheExpiration).Err(); err != nil {
		return nil, fmt.Errorf("failed to save client rand in cache")
	}

	return &powproto.Message{
		Kind:    powproto.ChallengeResponse,
		Payload: hc.String(),
	}, nil
}

func (s *Server) processQuote(ctx context.Context, requestData string, clientAddr net.Addr) (*powproto.Message, error) {
	hc, err := hashcash.Unmarshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request data: %w", err)
	}

	if err := hc.Verify(); err != nil {
		return nil, fmt.Errorf("failed to verify challenge: %w", err)
	}

	cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, clientAddr.String())
	rnd, err := s.cacheClient.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read client rand from cache")
	}
	if rnd != hc.Rand {
		return nil, ErrInvalidChallengeRand
	}

	if err := s.cacheClient.Del(ctx, cacheKey).Err(); err != nil {
		return nil, fmt.Errorf("failed to delete client rand from cache")
	}

	rndIdx := rand.Intn(len(quotes))
	return &powproto.Message{
		Kind:    powproto.QuoteResponse,
		Payload: quotes[rndIdx],
	}, nil
}

func writeMessage(msg powproto.Message, conn net.Conn) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), powproto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
