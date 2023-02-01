package hashcash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	hashcashVersion       = 1
	hashcashBits          = 20
	hashcashItemCount     = 7
	hashcashRandomBytes   = 8
	hashcashMaxIterations = 1 << 32
	hashcashTimeFormat    = "060102150405" // YYMMDDhhmmss
	hashcashExpiration    = 1 * time.Minute

	bitsPerHex    = 4
	codePointZero = 48
)

var (
	ErrHashNotComputed    = errors.New("hashcash: hash not computed")
	ErrHashNotMatched     = errors.New("hashcash: hash not matched")
	ErrInvalidDate        = errors.New("hashcash: invalid date")
	ErrInvalidFormat      = errors.New("hashcash: invalid format")
	ErrUnsupportedVersion = errors.New("hashacash: unsupported version")
)

type Hashcash struct {
	Version   int
	Bits      int
	Date      time.Time
	Resource  string
	Extension string
	Rand      string
	Counter   int
}

func NewHashcash(resource string) (*Hashcash, error) {
	rand, err := randomBytes(hashcashRandomBytes)
	if err != nil {
		return nil, err
	}

	return &Hashcash{
		Version:   hashcashVersion,
		Bits:      hashcashBits,
		Date:      time.Now().UTC(),
		Resource:  resource,
		Extension: "",
		Rand:      base64EncodeBytes(rand),
		Counter:   1,
	}, nil
}

func (h *Hashcash) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%s:%s:%s", h.Version, h.Bits, h.Date.Format(hashcashTimeFormat),
		h.Resource, h.Extension, h.Rand, base64EncodeInt(h.Counter))
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func base64EncodeBytes(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func base64DecodeBytes(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func base64EncodeInt(n int) string {
	return base64EncodeBytes([]byte(strconv.Itoa(n)))
}

func base64DecodeInt(s string) (int, error) {
	bs, err := base64DecodeBytes(s)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(bs))
}

func (h *Hashcash) Compute() error {
	zeroCount := h.Bits / bitsPerHex
	hash := sha1Hash(h.String())

	for !isHashMatched(hash, codePointZero, zeroCount) {
		h.Counter++
		hash = sha1Hash(h.String())
		if h.Counter >= hashcashMaxIterations {
			return ErrHashNotComputed
		}
	}
	return nil
}

func sha1Hash(s string) string {
	h := sha1.New()
	_, err := io.WriteString(h, s)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isHashMatched(hash string, char rune, count int) bool {
	for _, c := range hash[:count] {
		if c != char {
			return false
		}
	}
	return true
}

func (h *Hashcash) Verify() error {
	zeroCount := h.Bits / bitsPerHex
	challenge := h.String()
	hash := sha1Hash(challenge)

	if !isHashMatched(hash, codePointZero, zeroCount) {
		return ErrHashNotMatched
	}

	now := time.Now().UTC()
	if h.Date.After(now) || now.Sub(h.Date) >= hashcashExpiration {
		return ErrInvalidDate
	}
	return nil
}

func Unmarshal(s string) (*Hashcash, error) {
	hcItems := strings.Split(s, ":")
	if len(hcItems) != hashcashItemCount {
		return nil, ErrInvalidFormat
	}

	ver, err := strconv.Atoi(hcItems[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version type: %w", err)
	}
	if ver != hashcashVersion {
		return nil, ErrUnsupportedVersion
	}

	bits, err := strconv.Atoi(hcItems[1])
	if err != nil {
		return nil, fmt.Errorf("invalid zeroes bits type: %w", err)
	}

	date, err := time.Parse(hashcashTimeFormat, hcItems[2])
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	counter, err := base64DecodeInt(hcItems[6])
	if err != nil {
		return nil, fmt.Errorf("invalid counter format: %w", err)
	}

	return &Hashcash{
		Version:   ver,
		Bits:      bits,
		Date:      date,
		Resource:  hcItems[3],
		Extension: hcItems[4],
		Rand:      hcItems[5],
		Counter:   counter,
	}, nil
}
