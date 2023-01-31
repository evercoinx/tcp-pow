package hashcash

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	hashcashVersion    = 1
	hashcashBits       = 5
	hashcashItemCount  = 7
	hashcashTimeFormat = "060102150405" // YYMMDDhhmmss
)

type Hashcash struct {
	version   int
	bits      int
	date      time.Time
	resource  string
	extension string
	rand      string
	counter   int
}

func NewHashcash(resource string) (Hashcash, error) {
	rand, err := randomBytes(8)
	if err != nil {
		return Hashcash{}, err
	}

	return Hashcash{
		version:   hashcashVersion,
		bits:      hashcashBits,
		date:      time.Now().UTC(),
		resource:  resource,
		extension: "",
		rand:      base64EncodeBytes(rand),
		counter:   1,
	}, nil
}

func (h Hashcash) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%s:%s:%s", h.version, h.bits, h.date.Format(hashcashTimeFormat),
		h.resource, h.extension, h.rand, base64EncodeInt(h.counter))
}

func Parse(s string) (*Hashcash, error) {
	hcItems := strings.Split(s, ":")
	if len(hcItems) != hashcashItemCount {
		return nil, errors.New("invalid hashcash string format")
	}

	ver, err := strconv.Atoi(hcItems[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version type: %w", err)
	}
	if ver != hashcashVersion {
		return nil, fmt.Errorf("unsupported version: %d", ver)
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
		return nil, fmt.Errorf("failed to decode counter: %w", err)
	}

	return &Hashcash{
		version:   ver,
		bits:      bits,
		date:      date,
		resource:  hcItems[3],
		extension: hcItems[4],
		rand:      hcItems[5],
		counter:   counter,
	}, nil
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
