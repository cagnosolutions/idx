package main

import (
	"fmt"
	"testing"
)

func Benchmark_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		tree.Add([]byte(k), i)
	}
}

func Benchmark_Set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		tree.Set([]byte(k), i)
	}
}

func Benchmark_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		tree.Get([]byte(k))
	}
}
