package tcpserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/evercoinx/tcp-pow-server/internal/hashcash"
	"github.com/evercoinx/tcp-pow-server/internal/powproto"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const cacheKeyPrefix = "tcpserver:"

var (
	ErrUnexpectedMessageKind  = errors.New("unexpected message kind")
	ErrUnsupportedMessageKind = errors.New("unsupported message kind")
	ErrInvalidChallengeRand   = errors.New("invalid challenge rand")
)

var wisdomQuotes = []string{
	"The journey of a thousand miles begins with one step.",
	"He is no fool who gives what he cannot keep to gain what he cannot lose.",
	"It's not what you look at that matters, it's what you see.",
	"A man should always consider how much he has more than he wants.",
	"That old law about 'an eye for an eye' leaves everybody blind. The time is always right to do the right thing.",
}

type Server struct {
	cache           *redis.Client
	cacheExpiration time.Duration
}

func NewServer(rc *redis.Client, cacheExpiration time.Duration) *Server {
	return &Server{
		cache:           rc,
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
	resource := strings.Replace(clientAddr.String(), ":", "/", 1)
	hc, err := hashcash.NewHashcash(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	cacheKey := fmt.Sprintf("%s%s", cacheKeyPrefix, clientAddr.String())
	if err := s.cache.SetEx(ctx, cacheKey, hc.Rand, s.cacheExpiration).Err(); err != nil {
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
	rnd, err := s.cache.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read client rand from cache")
	}
	if rnd != hc.Rand {
		return nil, ErrInvalidChallengeRand
	}

	if err := s.cache.Del(ctx, cacheKey).Err(); err != nil {
		return nil, fmt.Errorf("failed to delete client rand from cache")
	}

	quote := wisdomQuotes[rand.Intn(len(wisdomQuotes))]
	return &powproto.Message{
		Kind:    powproto.QuoteResponse,
		Payload: quote,
	}, nil
}

func writeMessage(msg powproto.Message, conn net.Conn) error {
	rawMsg := fmt.Sprintf("%s%c", msg.String(), powproto.MessageTerminator)
	_, err := conn.Write([]byte(rawMsg))
	return err
}
