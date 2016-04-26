package idx

import (
	"encoding/json"
	"sync"
)

type StorageEngine interface {
	Has(offset int) bool
	Set(offset int, data []byte)
	Get(offset int) []byte
	Del(offset int)
}

type Engine struct {
	path   string
	bitmap *BitMmap
	datmap *DatMmap
	sync.RWMutex
}

// open a storage engine, or create if needed and align the
// size to the minimum memory mapped file size (ie. 16 MB)
func OpenEngine(path string) *Engine {
	bitmap := OpenBitMmap(path)
	datmap := OpenDatMmap(path, bitmap.used)
	engine := &Engine{
		bitmap: bitmap,
		datmap: datmap,
	}
	return engine
}

// returns boolean indicating if there is data at given offset
func (e *Engine) Has(offset int) bool {
	e.RLock()
	defer e.RUnlock()
	return e.bitmap.Has(offset)
}

// updates existing or inserts new block at offset
func (e *Engine) Set(offset int, data []byte) {
	e.Lock()
	e.bitmap.Set(offset)
	e.datmap.Set(offset, data)
	e.Unlock()
}

// returns block at offset n
func (e *Engine) Get(offset int) []byte {
	e.RLock()
	defer e.RUnlock()
	if !e.bitmap.Has(offset) {
		return nil
	}
	return strip(e.datmap.Get(offset))
}

// removes block at offset
func (e *Engine) Del(offset int) {
	e.Lock()
	e.bitmap.Del(offset)
	e.datmap.Del(offset)
	e.Unlock()
}

type Record struct {
	Key []byte
	Val int
}

func (e *Engine) All() <-chan Record {
	e.RLock()
	defer e.RUnlock()
	var record []interface{}
	ch := make(chan Record)
	go func() {
		for _, page := range e.bitmap.All() {
			err := json.Unmarshal(e.datmap.Get(page), &record)
			if err != nil {
				panic(err)
			}
			ch <- Record{[]byte(record[0].(string)), page}
		}
		close(ch)
	}()
	return ch
}

// closes the mapped file
func (e *Engine) CloseEngine() {
	e.bitmap.CloseBitMmap()
	e.datmap.CloseDatMmap()
}

// open file helper
/*func OpenFile(path string) (*os.File, string, int) {
	fd, err := os.OpenFile(path, syscall.O_RDWR|syscall.O_CREAT|syscall.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	fi, err := fd.Stat()
	if err != nil {
		panic(err)
	}
	return fd, sanitize(fi.Name()), int(fi.Size())
}*/

/*// round up to nearest pagesize -- helper
func align(size int) int {
	if size > 0 {
		return (size + SYS_PAGE - 1) &^ (SYS_PAGE - 1)
	}
	return SYS_PAGE
}

// resize underlying file -- helper
func resize(fd uintptr, size int) int {
	err := syscall.Ftruncate(int(fd), int64(align(size)))
	if err != nil {
		panic(err)
	}
	return size
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
}*/
