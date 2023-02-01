package hashcash

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
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
	ErrInvalidHash        = errors.New("hashcash: invalid hash")
	ErrInvalidDate        = errors.New("hashcash: invalid date")
	ErrInvalidResource    = errors.New("hashcash: invalid resource")
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
		Rand:      encodeBase64Bytes(rand),
		Counter:   1,
	}, nil
}

func (h *Hashcash) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%s:%s:%s", h.Version, h.Bits, h.Date.Format(hashcashTimeFormat),
		url.QueryEscape(h.Resource), h.Extension, h.Rand, encodeBase64Int(h.Counter))
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func encodeBase64Bytes(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64Bytes(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func encodeBase64Int(n int) string {
	return encodeBase64Bytes([]byte(strconv.Itoa(n)))
}

func decodeBase64Int(s string) (int, error) {
	bs, err := decodeBase64Bytes(s)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(bs))
}

func (h *Hashcash) Compute() error {
	zeroCount := h.Bits / bitsPerHex
	hash := sha1Hash(h.String())

	for !isHashValid(hash, codePointZero, zeroCount) {
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

func isHashValid(hash string, char rune, count int) bool {
	for _, c := range hash[:count] {
		if c != char {
			return false
		}
	}
	return true
}

func (h *Hashcash) Verify() error {
	hash := sha1Hash(h.String())
	zeroCount := h.Bits / bitsPerHex
	if !isHashValid(hash, codePointZero, zeroCount) {
		return ErrInvalidHash
	}

	now := time.Now().UTC()
	if h.Date.After(now) || now.Sub(h.Date) >= hashcashExpiration {
		return ErrInvalidDate
	}

	rsItems := strings.Split(h.Resource, ":")
	if len(rsItems) != 2 || net.ParseIP(rsItems[0]) == nil {
		return ErrInvalidResource
	}
	port, err := strconv.Atoi(rsItems[1])
	if err != nil || port <= 0 || port >= 1<<16-1 {
		return ErrInvalidResource
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

	resource, err := url.QueryUnescape(hcItems[3])
	if err != nil {
		return nil, fmt.Errorf("invalid resource format: %w", err)
	}

	counter, err := decodeBase64Int(hcItems[6])
	if err != nil {
		return nil, fmt.Errorf("invalid counter format: %w", err)
	}

	return &Hashcash{
		Version:   ver,
		Bits:      bits,
		Date:      date,
		Resource:  resource,
		Extension: hcItems[4],
		Rand:      hcItems[5],
		Counter:   counter,
	}, nil
}
