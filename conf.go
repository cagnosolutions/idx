package idx

import "os"

// System's Virtual Memory Page size.
// Everything is based off of this, so
// if you would like to augment anything
// you will beed to change this. This is
// usually going to be a 4KB pagesize.
var SYS_PAGE = os.Getpagesize()

// ORDER is the maximum number of
// children a non-leaf node can hold
//const ORDER = 32

// Key and Val are the types that
// the b+tree holds in the nodes
// and leaf records
//type Key []byte
//type Val int

// Compare is the main comparitor
// function used by the b+tree
//func Compare(a, b Key) int {
//	return bytes.Compare(a, b)
//}
