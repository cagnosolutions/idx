package idx

import (
	"bytes"
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
	primary *Tree
	engine  *Engine
	indexes map[string]*Tree
	sync.RWMutex
}

func NewStore(path string, indexes ...string) *Store {
	st := &Store{}
	st.primary = NewTree()
	st.engine = OpenEngine(path)
	st.indexes = make(map[string]*Tree, 0)
	for _, index := range indexes {
		st.indexes[index] = NewTree()
	}
	for key, page := range st.engine.All() {
		st.primary.Set(key, page)
	}
	return st
}

func (st *Store) Add(k []byte, v interface{}) error {
	st.Lock()
	defer st.Unlock()
	if !st.primary.Has(k) {
		doc, err := encode(k, v)
		if err != nil {
			return err
		}
		page := st.metamap.Add()
		if page == -1 {
			return ErrStoreFull
		}
		st.filemap.Set(page, doc)
		st.primary.Set(k, page)
		return nil
	}
	return ErrExists
}

func (st *Store) Set(k []byte, v interface{}) error {
	st.Lock()
	defer st.Unlock()
	doc, err := encode(k, v)
	if err != nil {
		return err
	}
	if v := st.primary.Get(k); v != nil {
		st.filemap.Set(v, doc)
		return nil
	}
	page := st.metamap.Add()
	if page == -1 {
		return ErrStoreFull
	}
	st.filemap.Set(page, doc)
	st.primary.Set(k, page)
	return nil
}

func (st *Store) Get(k []byte, ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	if v := st.primary.Get(k); v != nil {
		if err := decode(v, ptr); err != nil {
			return err
		}
		return nil
	}
	return ErrNotFound
}

func (st *Store) Del(k []byte) {
	st.Lock()
	st.primary.Del(k)
	st.Unlock()
}

func (st *Store) All(ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	records := bytes.Join([][]byte(st.primary.All()), []byte{','})
	records = append([]byte{'['}, append(records, byte(']'))...)
	if err := decode(records, ptr); err != nil {
		return err
	}
	return nil
}
