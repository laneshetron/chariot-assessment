package id

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"
)

type ID [11]byte

var (
	last    uint32
	counter uint16
)

// New generates a cryptographically secure, Base32-encoded ID
func New() (ID, error) {
	combinedBytes := ID{}

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

func (id ID) String() string {
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id[:])

	return s[:7] + "-" + s[7:]
}
