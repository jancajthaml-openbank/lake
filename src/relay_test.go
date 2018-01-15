package main

import (
	"fmt"
	"testing"
)

func init() {
	// Setup ZMQ relay here and share it between tests
	go StartRelay()
}

func TestRelay(t *testing.T) {
	fmt.Println("Test Not implemented")
}

func BenchmarkRelay(b *testing.B) {
	fmt.Println("Benchmark Not implemented")
	b.SetBytes(16)

	for n := 0; n < b.N; n++ {
		fmt.Sprintf("Not implemented")
	}
}
