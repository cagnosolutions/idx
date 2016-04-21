package adb

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"syscall"
)

// bpt.go -- find first leaf
func findFirstLeaf(root *node) *node {
	if root == nil {
		return root
	}
	c := root
	for !c.isLeaf {
		c = c.ptrs[0].(*node)
	}
	return c
}

// bpt.go -- cut leaf in half
func cut(length int) int {
	if length%2 == 0 {
		return length / 2
	}
	return length/2 + 1
}

// file.go, meta.go -- open file helper
func OpenFile(path string) (*os.File, string, int) {
	fd, err := os.OpenFile(path, syscall.O_RDWR|syscall.O_CREAT|syscall.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	fi, err := fd.Stat()
	if err != nil {
		panic(err)
	}
	return fd, sanitize(fi.Name()), int(fi.Size())
}

// file.go, meta.go -- create nested directories if they don't exist
func mkdirs(path string) {
	_, err := os.Stat(path)
	if err != nil {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
	}
}

// store.go -- sanitize path
func sanitize(path string) string {
	if path[len(path)-1] == '/' {
		return path[:len(path)-1]
	}
	if x := strings.Index(path, "."); x != -1 {
		return path[:x]
	}
	return path
}

// file.go, meta.go (mmap) -- round up to nearest pagesize
func align(size int) int {
	if size > 0 {
		return (size + SYS_PAGE - 1) &^ (SYS_PAGE - 1)
	}
	return SYS_PAGE
}

// file.go, meta.go (mmap) -- resize underlying file
func resize(fd uintptr, size int) int {
	err := syscall.Ftruncate(int(fd), int64(align(size)))
	if err != nil {
		panic(err)
	}
	return size
}

// bpt.go, file.go -- strip null bytes out of page
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

// store.go -- encode into a document
func encode(k string, v interface{}) ([]byte, error) {
	data := []interface{}{k, v}
	b, err := json.Marshal(data)
	if err != nil {
		Logger(err.Error())
		return nil, err
	}
	if len(b) > SYS_PAGE {
		Logger(ErrTooLarge.Error())
		return nil, ErrTooLarge
	}
	return b, nil
}

// store.go -- decode doc into a pointer supplied by the user
func decode(b []byte, v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		Logger(ErrNonPtrVal.Error())
		return ErrNonPtrVal
	}
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	return nil
}

// bpt.go, file.go -- return document value from page
func getdoc(b []byte, klen int) []byte {
	for i, j, set := klen+4, len(b)-1, 1; i < j; i, j = i+1, j-1 {
		if b[i] == '[' {
			set++
		}
		if b[i] == ']' {
			set--
		}
		if set == 0 || b[j] == ']' {
			if b[i] == ']' {
				return b[klen+4 : i]
			}
			return b[klen+4 : j]
		}
	}
	return b
}

// just a quick hack
func docbody(d []byte) []byte {
	var idx int
	var beg bool
	for i, b := range d {
		if b == '[' && i == 0 {
			beg = true
			idx++
		}
		if beg && b == ',' {
			break
		}
		idx++
	}
	return d[idx : len(d)-1]
}

func query(b []byte, qry ...string) bool {
	n := len(qry)
	for _, d := range bytes.FieldsFunc(bytes.Map(func(r rune) rune {
		switch r {
		case '{', '}', '[', ']', '"':
			return 0x00
		case ':':
			return '='
		}
		return r
	}, b), func(r rune) bool { return r == ',' }) {
		d = bytes.Replace(d, []byte{0x00}, []byte{}, -1)
		for i := 0; i < len(qry); i++ {
			q := []byte(qry[i])
			if bytes.Contains(q, []byte{'<'}) {
				dk, qk := bytes.Split(d, []byte{'='}), bytes.Split(q, []byte{'<'})
				if bytes.Equal(dk[0], qk[0]) {
					if len(dk[1]) < len(qk[1]) {
						n--
					}
					if len(dk[1]) == len(qk[1]) && bytes.Compare(dk[1], qk[1]) < 0 {
						n--
					}
				}
			}
			if bytes.Contains(q, []byte{'>'}) {
				dk, qk := bytes.Split(d, []byte{'='}), bytes.Split(q, []byte{'>'})
				if bytes.Equal(dk[0], qk[0]) {
					if len(dk[1]) > len(qk[1]) {
						n--
					}
					if len(dk[1]) == len(qk[1]) && bytes.Compare(dk[1], qk[1]) > 0 {
						n--
					}
				}
			}
			if bytes.Contains(q, []byte{'^'}) {
				dk, qk := bytes.Split(d, []byte{'='}), bytes.Split(q, []byte{'^'})
				if bytes.Equal(dk[0], qk[0]) && !bytes.Equal(dk[1], qk[1]) {
					n--
				}
			}
			if bytes.Contains(d, q) {
				n--
			}
		}
	}
	return n == 0
}
