package main

import "testing"

func BenchmarkGenerateID(b *testing.B) {
	id := ID{}
	for n := 0; n < b.N; n++ {
		id.Generate()
	}
}
