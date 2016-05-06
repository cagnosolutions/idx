package main

import "github.com/cagnosolutions/idx"

func main() {
	t := idx.NewTree()
	t.Set([]byte(`1`), 1)
	t.Set([]byte(`2`), 2)
	t.Set([]byte(`222`), 222)
	t.Set([]byte(`3`), 3)
	t.Set([]byte(`4`), 4)
	t.Set([]byte(`12`), 12)
	t.Set([]byte(`5`), 5)
	t.Set([]byte(`6`), 6)
	t.Set([]byte(`9`), 9)
	t.Set([]byte(`10`), 10)
	t.Set([]byte(`8`), 8)
	t.Set([]byte(`11`), 11)
	t.Set([]byte(`7`), 7)
	t.Set([]byte(`13`), 13)
	t.Set([]byte(`14`), 14)
	t.BFS()
}
