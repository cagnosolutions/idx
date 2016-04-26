package idx

import (
	"encoding/json"
	"os"
)

var (
	nilPage = make([]byte, SYS_PAGE)
	tbl     = [16]byte{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}
)

const (
	DATAOFFSET = 65536
)

type MappedData struct {
	path string
	file *os.File
	size int
	used int
	mmap Data
}

// open a mapped file, or create if needed and align the
// size to the minimum memory mapped file size (ie. 16 MB)
func OpenMappedData(path string) *MappedData {
	file, path, size := OpenFile(path + ".dat")
	if size == 0 {
		size = resize(file.Fd(), 1<<24) // start size 16MB
	}
	md := &MappedData{
		path: path + ".dat",
		file: file,
		size: size,
		mmap: Mmap(file, 0, size),
	}
	md.bitMapUsed()
	return md
}

// updates existing or inserts new block at offset n
func (md *MappedData) Add(b []byte) int {
	md.checkGrow()
	n := md.bitMapAdd()
	if n == -1 {
		return -1
	}
	// new position has been set in bitmap
	pos := getOffset(n)
	copy(md.mmap[pos:pos+SYS_PAGE], b)
	md.used++
	return n
}

// updates existing or inserts new block at offset n
func (md *MappedData) Set(n int, b []byte) {
	md.checkGrow()
	pos := getOffset(n)
	if !md.bitMapHas(n) {
		md.used++ // we are adding
		md.bitMapSet(n)
	} else {
		//copy(nilPage, b) // wipe existing record data
		copy(md.mmap[pos:pos+SYS_PAGE], nilPage)
	}
	// otherwise we are just updating
	copy(md.mmap[pos:pos+SYS_PAGE], b)
}

// returns block at offset n
func (md *MappedData) Get(n int) []byte {
	if md.bitMapHas(n) {
		pos := getOffset(n)
		return strip(md.mmap[pos : pos+SYS_PAGE])
	}
	return nil
}

// removes block at offset n
func (md *MappedData) Del(n int) {
	if md.bitMapHas(n) {
		md.bitMapDel(n)
		pos := getOffset(n)
		copy(md.mmap[pos:pos+SYS_PAGE], nilPage)
		md.used--
	}
}

func (md *MappedData) All() map[string]int {
	m := make(map[string]int)
	v := []interface{}{}
	for _, page := range md.bitMapAll() {
		b := md.Get(page)
		if err := json.Unmarshal(b, &v); err != nil {
			panic(err)
		}
		m[v[0].(string)] = page
	}
	return m
}

// closes the mapped file
func (md *MappedData) CloseMappedData() {
	md.mmap.Sync()
	md.mmap.Munmap()
	md.file.Close()
}

// check to see if we should grow
func (md *MappedData) checkGrow() {
	if md.used+1 < (md.size-DATAOFFSET)/SYS_PAGE {
		return // no need to grow
	}
	// unmap, grow underlying file and remap
	//md.mmap.Munmap()
	//md.size = resize(md.file.Fd(), md.size+(1<<24)) // grow size 16MB
	//md.mmap = Mmap(md.file, 0, md.size)

	md.mmap = md.mmap.Mremap(md.size + (1 << 24))
	md.size = md.size + (1 << 24)
}

func (md *MappedData) bitMapHas(k int) bool {
	return (md.mmap[k/8] & (1 << (uint(k % 8)))) != 0
}

func (md *MappedData) bitMapAdd() int {
	if k := md.bitMapNext(); k != -1 {
		md.bitMapSet(k) // add
		return k
	}
	return -1
}

func (md *MappedData) bitMapSet(k int) {
	// flip the n-th bit on; add/set
	md.mmap[k/8] |= (1 << uint(k%8))
}

func (md *MappedData) bitMapDel(k int) {
	// flip the k-th bit off; delete
	md.mmap[k/8] &= ^(1 << uint(k%8))
}

func (md *MappedData) bits(n byte) int {
	return int(tbl[n>>4] + tbl[n&0x0f])
}

func (md *MappedData) bitMapNext() int {
	for i := 0; i < (524272 / 8); i++ {
		if md.bits(md.mmap[i]) < 8 {
			for j := 0; j < 8; j++ {
				cur := (i * 8) + j
				if !md.bitMapHas(cur) {
					return cur
				}
			}
		}
	}
	return -1
}

func (md *MappedData) bitMapUsed() {
	for i := 0; i < (524272 / 8); i++ {
		md.used += md.bits(md.mmap[i])
	}
}

func (md *MappedData) bitMapAll() []int {
	var all []int
	for i := 0; i < (524272 / 8); i++ {
		if md.bits(md.mmap[i]) <= 8 {
			for j := 0; j < 8; j++ {
				cur := (i * 8) + j
				if md.bitMapHas(cur) {
					all = append(all, cur)
				}
			}
		}
	}
	return all
}

func getOffset(pos int) int {
	return (pos * SYS_PAGE) + DATAOFFSET
}

// strip null bytes out of page
func strip(b []byte) []byte {
	for i, j := 0, len(b)-1; i <= j; i, j = i+1, j-1 {
		if b[i] == 0x00 {
			return b[:i]
		}
		if b[j] != 0x00 {
			return b[:j+1]
		}
	}
	return b
}
