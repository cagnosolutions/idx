package idx

import "os"

// System's Virtual Memory Page size.
// Everything is based off of this, so
// if you would like to augment anything
// you will beed to change this. This is
// usually going to be a 4KB pagesize.
var SYS_PAGE = os.Getpagesize()

type DB interface {

	// checks for store existance
	HasStore(name string) bool

	// adds a new store if it doesn't exist
	AddStore(name string)

	// delete a store if it exists
	DelStore(name string)

	// update / change the store's name
	UpdStore(oldName, newName string) error

	// checks to see if a key value pair exists
	Has(k string) bool

	// adds a new key value pair if it doesn't exist
	Add(k string, v interface{})

	// updates an existing or sets a new key value pair (volitile)
	Set(k string, v interface{})

	// returns a records value using the supplied key
	Get(k string, v interface{}) error

	// deletes a key value pair if it exists
	Del(k string)
}

type DBStore interface {

	// checks to see if a key value pair exists
	Has(k string) bool

	// adds a new key value pair if it doesn't exist
	Add(k string, v interface{})

	// updates an existing or sets a new key value pair (volitile)
	Set(k string, v interface{})

	// returns a records value using the supplied key
	Get(k string) interface{}

	// deletes a key value pair if it exists
	Del(k string)
}

type Engine interface {
}

type Index interface {

	// checks to see if a key value pair exists
	Has(k string) bool

	// adds a new key value pair if it doesn't exist
	Add(k string, v interface{})

	// updates an existing or sets a new key value pair (volitile)
	Set(k string, v interface{})

	// returns a records value using the supplied key
	Get(k string) interface{}

	// deletes a key value pair if it exists
	Del(k string)
}
