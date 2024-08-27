package id

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ID [11]byte

var (
	last    uint32
	counter uint16
)

// New generates a cryptographically secure, Base32-encoded ID
func New() (ID, error) {
	var combinedBytes ID

	// seconds since Jan 1 2020
	now := uint32(time.Now().Unix() - 1577854800)
	binary.BigEndian.PutUint32(combinedBytes[:4], now)
	// reset counter every second
	if last != now {
		last = now
		counter = 0
	}

	// 16-bit counter
	binary.BigEndian.PutUint16(combinedBytes[4:6], uint16(counter))
	counter++

	// 40-bit random string, ~1.1 trillion values
	if _, err := rand.Read(combinedBytes[6:]); err != nil {
		return combinedBytes, fmt.Errorf("failed to generate random bytes: %v", err)
	}
	return combinedBytes, nil
}

func FromString(s string) (ID, error) {
	var id ID
	switch len(s) {
	case 17:
		// without dash
		s = strings.ToUpper(s)
	case 18:
		// with dash
		s = strings.ToUpper(s)
		s = s[:7] + s[8:]
	default:
		return id, errors.New("Invalid ID: must be 17 or 18 characters")
	}
	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		return id, errors.New("Invalid ID: could not decode")
	}
	copy(id[:], data)
	return id, nil
}

func FromBytes(b []byte) (ID, error) {
	var id ID
	if len(b) != 11 {
		return id, errors.New("Invalid ID")
	}
	copy(id[:], b)
	return id, nil
}

func (id ID) String() string {
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:])

	return s[:7] + "-" + s[7:]
}
