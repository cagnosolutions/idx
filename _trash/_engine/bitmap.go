package engine

import "os"

const (
	metaSize = 1 << 19 // in bits
	wordSize = 8       // bits / word
)

var tbl = [16]byte{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

type MappedMeta struct {
	path string
	file *os.File
	size int
	used int
	meta Data
}

func OpenMappedMeta(path string) *MappedMeta {
	file, path, size := OpenFile(path + ".idx")
	if size == 0 {
		size = resize(file.Fd(), metaSize/wordSize)
	}
	mx := &MappedMeta{}
	mx.path = path + ".idx"
	mx.file = file
	mx.size = size
	mx.meta = Mmap(file, 0, size)
	mx.used = mx.Used()
	return mx
}

func (mx *MappedMeta) Has(k int) bool {
	return (mx.meta[k/wordSize] & (1 << (uint(k % wordSize)))) != 0
}

func (mx *MappedMeta) Add() int {
	if k := mx.Next(); k != -1 { // NOTE: this should never be -1
		mx.Set(k) // add
		return k
	}
	return -1
}

func (mx *MappedMeta) Set(k int) {
	// flip the n-th bit on; add/set
	mx.meta[k/wordSize] |= (1 << uint(k%wordSize))
	mx.used++
}

func (mx *MappedMeta) Del(k int) {
	// flip the k-th bit off; delete
	mx.meta[k/wordSize] &= ^(1 << uint(k%wordSize))
	mx.used--
}

func (mx *MappedMeta) bits(n byte) int {
	return int(tbl[n>>4] + tbl[n&0x0f])
}

// closes the mapped file
func (mx *MappedMeta) CloseMappedMeta() {
	mx.meta.Sync()
	mx.meta.Munmap()
	mx.file.Close()
}

func (mx *MappedMeta) Next() int {
	mx.checkGrow()
	for i := 0; i < len(mx.meta); i++ {
		if mx.bits(mx.meta[i]) < 8 {
			for j := 0; j < 8; j++ {
				cur := (i * wordSize) + j
				if !mx.Has(cur) {
					return cur
				}
			}
		}
	}
	return -1 // NOTE: this should be unreachable: grow failed
}

func (mx *MappedMeta) Used() int {
	var used int
	for i := 0; i < mx.size; i++ {
		used += mx.bits(mx.meta[i])
	}
	return used
}

func (mx *MappedMeta) All() []int {
	var all []int
	for i := 0; i < len(mx.meta); i++ {
		if mx.bits(mx.meta[i]) <= 8 {
			for j := 0; j < 8; j++ {
				cur := (i * wordSize) + j
				if mx.Has(cur) {
					all = append(all, cur)
				}
			}
		}
	}
	return all
}

func (mx *MappedMeta) checkGrow() {
	if mx.used+1 < mx.size*8 {
		return // no need to grow
	}
	mx.meta.Munmap()
	mx.size = resize(mx.file.Fd(), mx.size+(metaSize/wordSize))
	mx.meta = Mmap(mx.file, 0, mx.size)
}
