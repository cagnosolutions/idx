package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cagnosolutions/adb"
)

var (
	COUNT = 10
	STORE = "users"
)

type User struct {
	Name   string
	Age    int
	Active bool
	Addrs  []Address
}

type Address struct {
	Name   string
	Street string
	City   string
	State  string
	Zip    string
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

	fmt.Printf("adding %d records to store %q...\n", COUNT, STORE)
	sleep(2)

	for i := 0; i < COUNT; i++ {
		k := fmt.Sprintf("%d-%d", i, i)
		v := MakeUser(i)
		db.Add(STORE, k, v)
	}

	fmt.Printf("done adding records; let's get ALL records...\n")
	var users []User
	if ok := db.All(STORE, &users); !ok {
		log.Fatalf("Error getting all users from %s\n", STORE)
	}

	fmt.Printf("printing out individual users found...\n")
	for i, u := range users {
		fmt.Printf("(%d) %+v\n", i, u)
	}

	fmt.Printf("\nfinished...\n")
	// close
	db.Close()

	// wait... press any key to continue
	pause()

}

func MakeUser(i int) User {
	var addresses []Address
	address1 := Address{
		fmt.Sprintf("Address %d", i),
		fmt.Sprintf("%d23 Main Street", i+1),
		fmt.Sprintf("City %d", i),
		fmt.Sprintf("State %d", i),
		fmt.Sprintf("1234%d", i),
	}

	address2 := Address{
		fmt.Sprintf("Address %d", i+1),
		fmt.Sprintf("%d23 Main Street", i+2),
		fmt.Sprintf("City %d", i+1),
		fmt.Sprintf("State %d", i+1),
		fmt.Sprintf("1234%d", i+1),
	}
	addresses = append(addresses, address1, address2)
	return User{
		fmt.Sprintf("Name %d", i),
		i,
		i%2 == 0,
		addresses,
	}
}

func pause() {
	var n int
	fmt.Println("Press any key to continue...")
	fmt.Scanln(&n)
}
