package powproto_test

import (
	"testing"

	"github.com/evercoinx/tcp-pow/internal/powproto"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	t.Run("empty payload", func(t *testing.T) {
		kind := powproto.ChallengeRequest
		payload := ""

		m := powproto.NewMessage(kind, payload)

		require.Equal(t, m, powproto.Message{
			Kind:    kind,
			Payload: payload,
		})
	})

	t.Run("nonempty payload", func(t *testing.T) {
		kind := powproto.QuoteRequest
		payload := "1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ=="

		m := powproto.NewMessage(kind, payload)

		require.Equal(t, m, powproto.Message{
			Kind:    kind,
			Payload: payload,
		})
	})
}

func TestString(t *testing.T) {
	m := powproto.Message{
		Kind:    powproto.QuoteRequest,
		Payload: "1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==",
	}

	s := m.String()

	require.Equal(t, s, "21:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")
}

func TestParse(t *testing.T) {
	t.Run("success with empty payload", func(t *testing.T) {
		m, err := powproto.Parse("0")

		require.NoError(t, err)
		require.Equal(t, m, &powproto.Message{
			Kind:    powproto.ChallengeRequest,
			Payload: "",
		})
	})

	t.Run("success with nonempty payload", func(t *testing.T) {
		m, err := powproto.Parse("21:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.NoError(t, err)
		require.Equal(t, m, &powproto.Message{
			Kind:    powproto.QuoteRequest,
			Payload: "1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==",
		})
	})

	t.Run("error with zero length", func(t *testing.T) {
		m, err := powproto.Parse("")

		require.Equal(t, err, powproto.ErrZeroLengthData)
		require.Nil(t, m)
	})

	t.Run("error with invalid message kind", func(t *testing.T) {
		m, err := powproto.Parse("5")

		require.Equal(t, err, powproto.ErrInvalidMessageKind)
		require.Nil(t, m)
	})
}
