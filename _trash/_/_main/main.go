package main

import (
	"fmt"
	"time"

	"github.com/cagnosolutions/adb"
)

var (
	COUNT = 524289
	STORE = "users"
)

type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Active bool   `json:"active"`
}

func sleep(n int) {
	time.Sleep(time.Duration(n) * time.Second)
}

func main() {

	// create a new db instance
	fmt.Printf("creating a new store instance...\n")
	db := adb.NewDB()

	// add a new store (if id doesn't already exist....)
	db.AddStore(STORE)
	sleep(2)

	// add COUNT records to new store....
	fmt.Printf("adding %d records to store %q...\n", COUNT, STORE)
	sleep(2)

	for i := 0; i < COUNT; i++ {
		k := fmt.Sprintf("u-%.6d", i)
		v := User{fmt.Sprintf("User #%.6d", i), i, i%2 == 0}
		db.Add(STORE, k, v)
	}
	fmt.Printf("done adding records...\n")

	// range all records in order
	/*for _, r := range t.All() {
		fmt.Printf("doc-> k:%x, v:%s\n", r.Key, r.Val)
	}*/

	// close
	db.Close()

	// wait... press any key to continue
	pause()
}

func pause() {
	var n int
	fmt.Println("Press any key to continue...")
	fmt.Scanln(&n)
}
