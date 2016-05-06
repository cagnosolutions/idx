package main

import (
	"fmt"

	"github.com/cagnosolutions/idx/btree"
)

var COUNT = 500000

func main() {
	t := btree.NewBTree()
	for i := 0; i < COUNT; i++ {
		t.Set(btree.KEY(fmt.Sprintf("key-%d", i)), i+100)
	}
	for i := 0; i < COUNT; i++ {
		r := t.Get(btree.KEY(fmt.Sprintf("key-%d", i)))
		if r != nil {
			fmt.Printf("[k:%s, v:%d]\n", r.Key, r.Val)
		}
	}

	fmt.Println("Press any key to continue...")
	var ln int
	fmt.Scanln(&ln)
	t.Close()
}
