package main

import (
	"fmt"
	"testing"

	"github.com/cagnosolutions/idx"
)

func Benchmark_Add(b *testing.B) {
	tree := idx.NewTree()
	for i := 0; i < b.N; i++ {
		k, v := fmt.Sprintf("key-%.5d", i), fmt.Sprintf("value-%.5d", i)
		tree.Add([]byte(k), []byte(v))
	}
}

func Benchmark_Put(b *testing.B) {
	tree := idx.NewTree()
	for i := 0; i < b.N; i++ {
		k, v := fmt.Sprintf("key-%.5d", i), fmt.Sprintf("value-%.5d", i)
		tree.Put([]byte(k), []byte(v))
	}
}

func Benchmark_Set(b *testing.B) {
	tree := idx.NewTree()
	for i := 0; i < b.N; i++ {
		k, v := fmt.Sprintf("key-%.5d", i), fmt.Sprintf("value-%.5d", i)
		tree.Set([]byte(k), []byte(v))
	}
}