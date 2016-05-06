package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cagnosolutions/idx"
)

const (
	BLOCK_SIZE   = 4096
	RECORD_COUNT = 524288
)

var SIZE = (BLOCK_SIZE * RECORD_COUNT) * 3

func main() {
	fd, _, _ := idx.OpenFile("test1.dat")
	//fd.Truncate(int64(SIZE))
	mmap := idx.Mmap(fd, 0, SIZE)
	//for i := 0; i < SIZE-BLOCK_SIZE; i += BLOCK_SIZE {
	//	copy(mmap[i:i+1], []byte{0xff})
	//}
	t1 := time.Now().UnixNano()
	for i := 0; i < SIZE; i += BLOCK_SIZE {
		if mmap[i] != 0x00 {
			log.Printf("Found something at block %d!\n", i/BLOCK_SIZE)
			break
		}
	}
	t2 := time.Now().UnixNano()
	fmt.Printf("%dns, %.2fmc, %.2fms, %.2fsec\n", (t2 - t1), float32(t2-t1)/1000, float32(t2-t1)/1000/1000, float32(t2-t1)/1000/1000/1000)

	var ln int
	fmt.Println("Press any key to continue...")
	fmt.Scanln(&ln)

	mmap.Munmap()
	if err := fd.Close(); err != nil {
		panic(err)
	}
}
