package main

import (
	"fmt"
	"time"
	"github.com/cagnosolutions/idx"
)

const N = 2500000

func MemoryTesting_Set() {
	tree := idx.NewTree()
	for i := 0; i < N; i++ {
		k, v := fmt.Sprintf("key-%.5d", i), fmt.Sprintf("value-%.5d", i)
		tree.Set([]byte(k), []byte(v))
	}
}

func main() {
	MemoryTesting_Set()
	fmt.Printf("Loaded tree with %d items, sleeping for 10 seconds...\n", N)
	time.Sleep(time.Duration(10)*time.Second)
}
