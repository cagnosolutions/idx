package main

import (
	"fmt"

	"github.com/cagnosolutions/adb"
)

type comp struct {
	fld, opt, val string
}

func main() {


	v := `select from users`
	s, f := adb.Q(v)
	fmt.Printf("query: %s\nstore: %s, fields: %s\n", v, s, f)

	v = `select from users name=scott, email="scottiecagno@gmail.com", age>28`
	s, f = adb.Q(v)
	fmt.Printf("query: %s\nstore: %s, fields: %s\n", v, s, f)

}
