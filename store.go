package idx

import (
	"errors"
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
		st.index.Set([]byte(key), page)
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
		st.index.Set(k, page)
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
