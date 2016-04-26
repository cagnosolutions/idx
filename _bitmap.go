package idx

import "os"

var tbl = [16]byte{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

type BitMmap struct {
	path string
	file *os.File
	size int
	used int
	mmap Data
}

func OpenBitMmap(path string) *BitMmap {
	file, path, size := OpenFile(path + ".idx")
	if size == 0 {
		// 524288 max number of indexable blocks relies on wordsize 8
		size = resize(file.Fd(), 524288/8)
	}

	bm := &BitMmap{}
	bm.path = path + ".idx"
	bm.file = file
	bm.size = size
	bm.mmap = Mmap(file, 0, size)
	bm.used = bm.Used()
	return bm
}

func (bm *BitMmap) Has(k int) bool {
	return (bm.mmap[k/8] & (1 << (uint(k % 8)))) != 0
}

func (bm *BitMmap) Add() int {
	if k := bm.Next(); k != -1 {
		bm.Set(k) // add
		return k
	}
	return -1
}

func (bm *BitMmap) Set(k int) {
	// flip the n-th bit on; add/set
	bm.mmap[k/8] |= (1 << uint(k%8))
	bm.used++
}

func (bm *BitMmap) Del(k int) {
	// flip the k-th bit off; delete
	bm.mmap[k/8] &= ^(1 << uint(k%8))
	bm.used--
}

func (bm *BitMmap) bits(n byte) int {
	return int(tbl[n>>4] + tbl[n&0x0f])
}

// closes the mapped file
func (bm *BitMmap) CloseBitMmap() {
	bm.mmap.Sync()
	bm.mmap.Munmap()
	bm.file.Close()
}

func (bm *BitMmap) Next() int {
	for i := 0; i < len(bm.mmap); i++ {
		if bm.bits(bm.mmap[i]) < 8 {
			for j := 0; j < 8; j++ {
				cur := (i * 8) + j
				if !bm.Has(cur) {
					return cur
				}
			}
		}
	}
	return -1
}

func (bm *BitMmap) Used() int {
	var used int
	for i := 0; i < bm.size; i++ {
		used += bm.bits(bm.mmap[i])
	}
	return used
}

func (bm *BitMmap) All() []int {
	var all []int
	for i := 0; i < len(bm.mmap); i++ {
		if bm.bits(bm.mmap[i]) <= 8 {
			for j := 0; j < 8; j++ {
				cur := (i * 8) + j
				if bm.Has(cur) {
					all = append(all, cur)
				}
			}
		}
	}
	return all
}
