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
	fmt.Printf("instantiating a db instance\nreading data from disk...\n")
	db := adb.NewDB()

	var users []User
	ok := db.All(STORE, &users)
	if !ok {
		log.Fatalf("ERROR GETTING USERS!!!!")
	}
	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}

	fmt.Printf("\ndone getting records...\n")
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
