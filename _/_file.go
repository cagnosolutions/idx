package adb

import (
	"os"
	"regexp"
)

var (
	nilPage = make([]byte, SYS_PAGE, SYS_PAGE)
)

const (
	fileSize = 1 << 24 // 16 MB
)

type MappedFile struct {
	path string
	file *os.File
	size int
	used int
	data Data
}

// open a mapped file, or create if needed and align the
// size to the minimum memory mapped file size (ie. 16 MB)
func OpenMappedFile(path string, used int) *MappedFile {
	file, path, size := OpenFile(path + ".dat")
	if size == 0 {
		size = resize(file.Fd(), fileSize)
	}
	data := Mmap(file, 0, size)
	return &MappedFile{
		path: path + ".dat",
		file: file,
		size: size,
		used: used,
		data: data,
	}
}

// updates existing or inserts new block at offset n
func (mf *MappedFile) Set(n int, b []byte) {
	mf.checkGrow()
	pos := n * SYS_PAGE
	if mf.data[pos] == 0x00 {
		mf.used++ // we are adding
	} else {
		//copy(nilPage, b) // wipe existing record data
		copy(mf.data[pos:pos+SYS_PAGE], nilPage)
	}
	// otherwise we are just updating
	copy(mf.data[pos:pos+SYS_PAGE], b)
}

// returns block at offset n
func (mf *MappedFile) Get(n int) []byte {
	pos := n * SYS_PAGE
	if n > -1 && mf.data[pos] != 0x00 {
		return strip(mf.data[pos : pos+SYS_PAGE])
	}
	return nil
}

// extracts and returns document from block at offset n
func (mf *MappedFile) GetDoc(n, kl int) []byte {
	pos := n * SYS_PAGE
	if n > -1 && mf.data[pos] != 0x00 {
		return getdoc(mf.data[pos:pos+SYS_PAGE], kl)
	}
	return nil
}

// extracts and returns document from block at offset n if it contains q
func (mf *MappedFile) GetDocMatch(n, kl int, qry *regexp.Regexp) []byte {
	pos := n * SYS_PAGE
	if n > -1 && mf.data[pos] != 0x00 {
		if qry.Match(mf.data[pos : pos+SYS_PAGE]) {
			return getdoc(mf.data[pos:pos+SYS_PAGE], kl)
		}
	}
	return nil
}

func (mf *MappedFile) GetDocQuery(n, kl int, qry ...string) []byte {
	pos := n * SYS_PAGE
	if n > -1 && mf.data[pos] != 0x00 {
		if query(mf.data[pos:pos+SYS_PAGE], qry...) {
			return getdoc(mf.data[pos:pos+SYS_PAGE], kl)
		}
	}
	return nil
}

func (mf *MappedFile) GetDocTestQuery(n, kl int, qry [][]byte) []byte {
	pos := n * SYS_PAGE
	if n > -1 && mf.data[pos] != 0x00 {
		match := true
		for _, q := range qry {
			if !check(q, mf.data[pos:pos+SYS_PAGE]) {
				match = false
				break
			}
		}
		if match {
			return getdoc(mf.data[pos:pos+SYS_PAGE], kl)
		}
	}
	return nil
}

// removes block at offset n
func (mf *MappedFile) Del(n int) {
	mf.used--
	pos := n * SYS_PAGE
	copy(mf.data[pos:pos+SYS_PAGE], nilPage)
}

// closes the mapped file
func (mf *MappedFile) CloseMappedFile() {
	mf.data.Sync()
	mf.data.Munmap()
	mf.file.Close()
}

// check to see if we should grow
func (mf *MappedFile) checkGrow() {
	if mf.used+1 < mf.size/SYS_PAGE {
		return // no need to grow
	}
	// unmap, grow underlying file and remap
	mf.data.Munmap()
	mf.size = resize(mf.file.Fd(), mf.size+fileSize)
	mf.data = Mmap(mf.file, 0, mf.size)
}
