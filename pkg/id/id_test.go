package id

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"
)

func BenchmarkNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		New()
	}
}

func TestFromStringDash(t *testing.T) {
	s := "BC5YUGA-aaTBBDEINA"
	_, err := FromString(s)
	if err != nil {
		t.Fatalf("Failed to parse string: got error %v", err)
	}
}

func TestFromStringNoDash(t *testing.T) {
	s := "BC5YUGAAATBBDEINA"
	_, err := FromString(s)
	if err != nil {
		t.Fatalf("Failed to parse string: got error %v", err)
	}
}

func TestFromStringInvalid(t *testing.T) {
	s := "BC5YUGAAATBBDEIN"
	_, err := FromString(s)
	if err == nil {
		t.Fatalf("got nil, want %v", errors.New("Invalid ID: must be 17 or 18 characters"))
	}
}

func TestMonotonic(t *testing.T) {
	var last ID
	for i := 0; i < 100; i++ {
		next, err := New()
		if err != nil {
			t.Fatalf("Failed to generate ID: got error %v", err)
		}
		if bytes.Compare(last[:], next[:]) >= 0 {
			t.Fatalf("Next ID not monotonic: (last, new) = (%v, %v)", last, next)
		}
		last = next
	}
}

func TestValidate(t *testing.T) {
	var id ID
	if valid, err := Validate(id); valid || err == nil {
		t.Fatal("Empty ID should be invalid")
	}
	id, _ = New()
	if valid, err := Validate(id); !valid || err != nil {
		t.Fatalf("Expected ID to be valid. want %v, nil got %v, %v", true, valid, err)
	}
	now := uint32(time.Now().Unix())
	binary.BigEndian.PutUint32(id[:4], now)
	if valid, err := Validate(id); valid || err == nil {
		t.Fatal("Expected ID with future timestamp to be invalid")
	}
}
