package main

import (
	"fmt"
	"testing"

	"github.com/cagnosolutions/idx"
)

func Benchmark_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := []byte(fmt.Sprintf("data-%.5d", i))
		tree.Add(x, x)
	}
}

func Benchmark_Set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := []byte(fmt.Sprintf("data-%.5d", i))
		tree.Set(x, x)
	}
}

func Benchmark_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := []byte(fmt.Sprintf("data-%.5d", i))
		tree.Get(x)
	}
}

var COUNT = 1000

var tree = idx.NewTree()

func TestSet(t *testing.T) {
	fmt.Println("Ran")
	for i := 0; i < COUNT; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		tree.Set([]byte(k), i)
	}
	if tree.Count() != COUNT {
		t.Errorf("tree.Count() != %d, it was %d", COUNT, tree.Count())
	}
}

func TestGet(t *testing.T) {
	for i := 0; i < COUNT; i++ {
		k := fmt.Sprintf("key-%.5d", i)
		r := tree.Get([]byte(k))
		if r != nil {
			if r.Val != i {
				t.Errorf("record val != %d, it was %d\n", i, r.Val)
			}
		} else {
			t.Errorf("record is nil, key is %s\n", k)
		}
	}
}
