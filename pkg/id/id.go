package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"
)

type ID struct {
	last    uint32
	counter uint16
}

// GenerateID generates a cryptographically secure, Base32-encoded ID
func (id *ID) Generate() (string, error) {
	combinedBytes := make([]byte, 11)

	// seconds since Jan 1 2020
	now := uint32(time.Now().Unix() - 1577854800)
	binary.BigEndian.PutUint32(combinedBytes[:4], now)
	// reset counter every second
	if id.last != now {
		id.last = now
		id.counter = 0
	}

	// 16-bit counter
	binary.BigEndian.PutUint16(combinedBytes[4:6], uint16(id.counter))
	id.counter++

	// 40-bit random string, ~1.1 trillion values
	if _, err := rand.Read(combinedBytes[6:]); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(combinedBytes)

	return s[:7] + "-" + s[7:], nil
}

func main() {
	// Generate and print an ID
	i := ID{}
	for {
		id, err := i.Generate()
		if err != nil {
			fmt.Printf("Error generating ID: %v\n", err)
			return
		}
		fmt.Print("Generated ID:", id, "\r")
	}
}
