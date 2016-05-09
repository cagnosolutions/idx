package main

import (
	"fmt"
	"testing"
)

//var tree = idx.NewTree()

func Benchmark_Set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		tree.Set([]byte(k), i)
	}
}

func Benchmark_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		if r := tree.Get([]byte(k)); r != nil {
			if r.Val != i {
				b.Fatalf("record val != %d\n", i)
			}
		} else {
			b.Fatalf("got nil record, KEY: %s\n", k)
		}
	}
}
