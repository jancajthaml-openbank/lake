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
	t.Log("Relays message")
	{
		fmt.Println("Test Not implemented")
	}
}

func BenchmarkRelay(b *testing.B) {
	b.SetBytes(16)
	for n := 0; n < b.N; n++ {

	}
}
