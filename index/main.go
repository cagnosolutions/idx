package main

import (
	"fmt"

	"github.com/cagnosolutions/idx"
)

func main() {
	t := idx.NewTree()
	t.Set([]byte(`001`), 1)
	t.Set([]byte(`002`), 2)
	t.Set([]byte(`003`), 3)
	t.Set([]byte(`004`), 4)
	t.Set([]byte(`012`), 12)
	t.Set([]byte(`005`), 5)
	t.Set([]byte(`006`), 6)
	t.Set([]byte(`009`), 9)
	t.Set([]byte(`010`), 10)
	t.Set([]byte(`008`), 8)
	t.Set([]byte(`011`), 11)
	t.Set([]byte(`007`), 7)
	t.Set([]byte(`013`), 13)
	t.Set([]byte(`014`), 14)
	fmt.Println(t.Count())
	fmt.Println(t.String())
}
