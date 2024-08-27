package id

import (
	"errors"
	"testing"
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
