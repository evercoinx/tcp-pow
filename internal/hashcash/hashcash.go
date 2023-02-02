// hashcash package implements a Hashcash and provides computation, verification and parsing
// operations for it.
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
	hashcashVersion       = 1               // Used protocol version
	hashcashBits          = 20              // Size of the `bits` field
	hashcashItemCount     = 7               // Total number of items ih a Hashcash string
	hashcashRandomBytes   = 8               // Bytes count as a source of entropy for the `rand` field
	hashcashMaxIterations = 1 << 32         // Max iterations to find a solution
	hashcashTimeFormat    = "060102150405"  // Time format of the `date` field as YYMMDDhhmmss
	hashcashExpiration    = 1 * time.Minute // Expiration interval for a Hashcash string

	bitsPerHex    = 4  // Number of bits per a hexidecimal character
	codePointZero = 48 // Unicode code point of character `0`
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
	Version   int       // Protocol version
	Bits      int       // Number of bits used to enforce challenge complexity
	Date      time.Time // Date of creation
	Resource  string    // Client's identification information as [host]:[port]
	Extension string    // Extension data that is empty for version 1
	Rand      string    // Random number to identify each Hashcash message
	Counter   int       // Attempt count when seeking a challenge solution
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

// String serializes the Hashcash into a string.
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

// Compute attempts to solve a challenge by incrementing the `counter` field in the Hashcash until
// it reaches the maximum iterations threshold.
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

// Verify carries out several tests to check the Hashcash.
func (h *Hashcash) Verify() error {
	// Test 1 checks that a computed Hashcash contains the required number of zeros.
	hash := sha1Hash(h.String())
	zeroCount := h.Bits / bitsPerHex
	if !isHashValid(hash, codePointZero, zeroCount) {
		return ErrInvalidHash
	}

	// Test 2 checks that the Hashcash `date` field is not in the future and is not too old.
	now := time.Now().UTC()
	if h.Date.After(now) || now.Sub(h.Date) >= hashcashExpiration {
		return ErrInvalidDate
	}

	// Test 3 checks that the Hashcash contains the valid `resource` field.
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

// Unmarshal deserializes a string into the Hashcash.
func Unmarshal(s string) (*Hashcash, error) {
	hcItems := strings.Split(s, ":")
	if len(hcItems) != hashcashItemCount {
		return nil, ErrInvalidFormat
	}

	ver, err := strconv.Atoi(hcItems[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}
	if ver != hashcashVersion {
		return nil, ErrUnsupportedVersion
	}

	bits, err := strconv.Atoi(hcItems[1])
	if err != nil {
		return nil, fmt.Errorf("invalid zero bits format: %w", err)
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
