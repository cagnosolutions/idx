package idx

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
)

/*
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
*/

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

/*
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
*/

// store.go -- encode into a document
func encode(k string, v interface{}) ([]byte, error) {
	data := []interface{}{k, v}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if len(b) > SYS_PAGE {
		return nil, ErrTooLarge
	}
	return b, nil
}

// store.go -- decode doc into a pointer supplied by the user
func decode(b []byte, v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
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
