package idx

import "bytes"

// ORDER is the maximum number of
// children a non-leaf node can hold
const ORDER = 32

// Key and Val are the types that
// the b+tree holds in the nodes
// and leaf records
type Key []byte
type Val []byte

// Compare is the main comparitor
// function used by the b+tree
func Compare(a, b Key) int {
	return bytes.Compare(a, b)
}

func Equal(a, b Key) bool {
	return bytes.Equal(a, b)
}
