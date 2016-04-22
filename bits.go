package idx

import (
	"log"
	"go/build"
)

// this is the word size of an int(32) 
// in bits. an int(32) requires a min-
// imum of 4 bytes, each of which are
// made up of 8 bits, therefore 4x8=32
// this same notion applies for int(64)
// such that 8 bytes * 8 bits/byte = 64
//
var WS = Arch()
var SZ = 1

type BitVec []int

func NewBitVec(n int) BitVec {
	if n > WS {
		SZ = (n/WS)+1
	}
	log.Printf("Bit vector of base size %d (%d max bits)\n", SZ, WS*SZ)
	return make([]int, SZ, SZ)
}

func Arch() int {
	if build.Default.GOARCH == "amd64" {
		return 64
	}
	return 32
}

func (bv BitVec) Has(k int) bool {
	return (bv[k/WS] & (1 << (uint(k % WS)))) != 0
}

func (bv BitVec) Add(k int) {
	bv[k/WS] |= (1 << uint(k % WS))
}

func (bv BitVec) Del(k int) {
	bv[k/WS] &= ^(1 << uint(k % WS))
}
