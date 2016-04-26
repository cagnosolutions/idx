package idx

import "os"

var (
	nilPage = make([]byte, SYS_PAGE, SYS_PAGE)
)

const (
	fileSize = 1 << 24 // 16 MB
)

type DatMmap struct {
	path string
	file *os.File
	size int
	used int
	mmap Data
}

// open a mapped file, or create if needed and align the
// size to the minimum memory mapped file size (ie. 16 MB)
func OpenDatMmap(path string, used int) *DatMmap {
	file, path, size := OpenFile(path + ".dat")
	if size == 0 {
		size = resize(file.Fd(), fileSize)
	}
	return &DatMmap{
		path: path + ".dat",
		file: file,
		size: size,
		used: used,
		mmap: Mmap(file, 0, size),
	}
}

// updates existing or inserts new block at offset n
func (dm *DatMmap) Set(n int, b []byte) {
	dm.checkGrow()
	pos := n * SYS_PAGE
	if dm.mmap[pos] == 0x00 {
		dm.used++ // we are adding
	} else {
		//copy(nilPage, b) // wipe existing record data
		copy(dm.mmap[pos:pos+SYS_PAGE], nilPage)
	}
	// otherwise we are just updating
	copy(dm.mmap[pos:pos+SYS_PAGE], b)
}

// returns block at offset n
func (dm *DatMmap) Get(n int) []byte {
	pos := n * SYS_PAGE
	if n > -1 && dm.mmap[pos] != 0x00 {
		return strip(dm.mmap[pos : pos+SYS_PAGE])
	}
	return nil
}

// extracts and returns document from block at offset n
func (dm *DatMmap) GetDoc(n, kl int) []byte {
	pos := n * SYS_PAGE
	if n > -1 && dm.mmap[pos] != 0x00 {
		return getdoc(dm.mmap[pos:pos+SYS_PAGE], kl)
	}
	return nil
}

// removes block at offset n
func (dm *DatMmap) Del(n int) {
	dm.used--
	pos := n * SYS_PAGE
	copy(dm.mmap[pos:pos+SYS_PAGE], nilPage)
}

// closes the mapped file
func (dm *DatMmap) CloseDatMmap() {
	dm.mmap.Sync()
	dm.mmap.Munmap()
	dm.file.Close()
}

// check to see if we should grow
func (dm *DatMmap) checkGrow() {
	if dm.used+1 < dm.size/SYS_PAGE {
		return // no need to grow
	}
	// unmap, grow underlying file and remap
	dm.mmap.Munmap()
	dm.size = resize(dm.file.Fd(), dm.size+fileSize)
	dm.mmap = Mmap(dm.file, 0, dm.size)
}
