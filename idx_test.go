package idx

import (
	"fmt"
	"sync"
	"testing"
)

var COUNT = 1000

var mu sync.RWMutex

var tree = NewTree()

func TestSet(t *testing.T) {
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
		mu.Lock()
		r := tree.Get([]byte(k))
		if r != nil {
			if r.Val != i {
				t.Errorf("record val != %d, it was %d\n", i, r.Val)
			}
		} else {
			t.Errorf("record is nil, key is %s\n", k)
		}
		mu.Unlock()
	}
}
