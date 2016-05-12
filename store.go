package idx

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
)

var (
	ErrTooLarge  = errors.New("key and value data is too large; maximum limit of 4KB")
	ErrStoreFull = errors.New("maximum number of records was reached; store is full")
	ErrNotFound  = errors.New("could not locate; not found")
	ErrNonPtrVal = errors.New("expected pointer to value, not value")
	ErrExists    = errors.New("key or value already exists")
)

type Store struct {
	index  *Tree
	engine *MappedData
	sync.RWMutex
}

func NewStore(path string) *Store {
	st := &Store{}
	st.index = NewTree()
	st.engine = OpenMappedData(path)
	for key, page := range st.engine.All() {
		st.index.Set([]byte(key), Itob(int64(page)))
	}
	return st
}

func (st *Store) Add(k []byte, v interface{}) error {
	st.Lock()
	defer st.Unlock()
	if !st.index.Has(k) {
		doc, err := encode(string(k), v)
		if err != nil {
			return err
		}
		page := st.engine.Add(doc)
		if page == -1 {
			return ErrStoreFull
		}
		st.index.Set(k, Itob(int64(page)))
		return nil
	}
	return ErrExists
}

func (st *Store) Set(k []byte, v interface{}) error {
	st.Lock()
	defer st.Unlock()
	doc, err := encode(string(k), v)
	if err != nil {
		return err
	}
	rec := st.index.Get(k)
	if rec != nil {
		st.engine.Set(rec.Val, doc)
		return nil
	}
	page := st.engine.Add(doc)
	if page == -1 {
		return ErrStoreFull
	}
	rec.Key = k
	rec.Val = page
	return nil
}

func (st *Store) Get(k []byte, ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	if r := st.index.Get(k); r != nil {
		if v := st.engine.Get(r.Val); v != nil {
			if err := decode(v, ptr); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrNotFound
}

func (st *Store) Del(k []byte) {
	st.Lock()
	st.index.Del(k)
	st.Unlock()
}

/*
func (st *Store) All(ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	records := bytes.Join([][]byte(st.index.All()), []byte{','})
	records = append([]byte{'['}, append(records, byte(']'))...)
	if err := decode(records, ptr); err != nil {
		return err
	}
	return nil
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
