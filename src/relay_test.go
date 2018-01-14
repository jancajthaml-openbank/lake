package main

import (
	"fmt"
	"testing"
)

func TestStub(t *testing.T) {
	fmt.Println("Test Not implemented")
}

func BenchmarkStub(b *testing.B) {
	fmt.Println("Benchmark Not implemented")
	b.SetBytes(16)

	for n := 0; n < b.N; n++ {
		fmt.Sprintf("Not implemented")
	}
}
