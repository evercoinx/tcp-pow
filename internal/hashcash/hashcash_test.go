package hashcash_test

import (
	"testing"
	"time"

	"github.com/evercoinx/tcp-pow/internal/hashcash"
	"github.com/stretchr/testify/require"
)

func TestNewHashcash(t *testing.T) {
	resource := "127.0.0.1:32000"

	h, err := hashcash.NewHashcash(resource)

	require.NoError(t, err)
	require.Equal(t, h.Version, 1)
	require.Equal(t, h.Bits, 20)
	require.NotEmpty(t, h.Date)
	require.Equal(t, h.Resource, resource)
	require.Empty(t, h.Extension)
	require.NotEmpty(t, h.Rand)
	require.Equal(t, h.Counter, 1)
}

func TestString(t *testing.T) {
	h := hashcash.Hashcash{
		Version:  1,
		Bits:     20,
		Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		Resource: "127.0.0.1:32000",
		Rand:     "Uv38ByGCZU8=",
		Counter:  1,
	}

	s := h.String()

	require.Equal(t, s, "1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")
}

func TestDecode(t *testing.T) {
	counter := 1472847
	h := hashcash.Hashcash{
		Version:  1,
		Bits:     20,
		Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		Resource: "127.0.0.1:32000",
		Rand:     "Uv38ByGCZU8=",
		Counter:  counter,
	}

	err := h.Compute()

	require.NoError(t, err)
	require.Equal(t, h.Counter, counter+1)
}

func TestVerify(t *testing.T) {
	t.Run("error with invalid hash", func(t *testing.T) {
		h := hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "127.0.0.1:32000",
			Rand:     "Uv38ByGCZU8=",
			Counter:  1472847,
		}

		err := h.Verify()

		require.Equal(t, err, hashcash.ErrInvalidHash)
	})

	t.Run("error with invalid date: too old", func(t *testing.T) {
		h := hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "127.0.0.1:32000",
			Rand:     "Uv38ByGCZU8=",
			Counter:  1472848,
		}

		err := h.Verify()

		require.Equal(t, err, hashcash.ErrInvalidDate)
	})

	t.Run("error with invalid date: in future", func(t *testing.T) {
		h := hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2038, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "127.0.0.1:32000",
			Rand:     "Uv38ByGCZU8=",
			Counter:  162408,
		}

		err := h.Verify()

		require.Equal(t, err, hashcash.ErrInvalidDate)
	})

	t.Run("error with invalid resource: ip address", func(t *testing.T) {
		h := hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "256.256.256.256:32000",
			Rand:     "Uv38ByGCZU8=",
			Counter:  2796951,
		}

		err := h.Verify()

		require.Equal(t, err, hashcash.ErrInvalidDate)
	})

	t.Run("error with invalid resource: port", func(t *testing.T) {
		h := hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "127.0.0.1:65536",
			Rand:     "Uv38ByGCZU8=",
			Counter:  5423364,
		}

		err := h.Verify()

		require.Equal(t, err, hashcash.ErrInvalidDate)
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("success with hashcash string", func(t *testing.T) {
		h, err := hashcash.Unmarshal("1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.NoError(t, err)
		require.Equal(t, h, &hashcash.Hashcash{
			Version:  1,
			Bits:     20,
			Date:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			Resource: "127.0.0.1:32000",
			Rand:     "Uv38ByGCZU8=",
			Counter:  1,
		})
	})

	t.Run("error with invalid version format", func(t *testing.T) {
		h, err := hashcash.Unmarshal("a:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.ErrorContains(t, err, "invalid version format")
		require.Nil(t, h)
	})

	t.Run("error with invalid zero bits format", func(t *testing.T) {
		h, err := hashcash.Unmarshal("1:a:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.ErrorContains(t, err, "invalid zero bits format")
		require.Nil(t, h)
	})

	t.Run("error with unsupported version", func(t *testing.T) {
		h, err := hashcash.Unmarshal("2:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.Equal(t, err, hashcash.ErrUnsupportedVersion)
		require.Nil(t, h)
	})

	t.Run("error with invalid date format", func(t *testing.T) {
		h, err := hashcash.Unmarshal("1:20:0601021504:127.0.0.1%3A32000::Uv38ByGCZU8=:MQ==")

		require.ErrorContains(t, err, "invalid date format")
		require.Nil(t, h)
	})

	t.Run("error with invalid resource format", func(t *testing.T) {
		h, err := hashcash.Unmarshal("1:20:060102150405:127.0.0.1%3G32000::Uv38ByGCZU8=:MQ==")

		require.ErrorContains(t, err, "invalid resource format")
		require.Nil(t, h)
	})

	t.Run("error with invalid counter format", func(t *testing.T) {
		h, err := hashcash.Unmarshal("1:20:060102150405:127.0.0.1%3A32000::Uv38ByGCZU8=:MQQ=")

		require.ErrorContains(t, err, "invalid counter format")
		require.Nil(t, h)
	})
}
