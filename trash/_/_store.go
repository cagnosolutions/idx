package adb

import (
	"bytes"
	"errors"
	"regexp"
	"sync"
)

var (
	ErrTooLarge  = errors.New("key and value data is too large; maximum limit of 4KB")
	ErrStoreFull = errors.New("maximum number of records was reached; store is full")
	ErrNotFound  = errors.New("could not locate; not found")
	ErrNonPtrVal = errors.New("expected pointer to value, not value")
)

type Store struct {
	index *Tree
	sync.RWMutex
}

func NewStore(path string) *Store {
	return &Store{
		index: NewTree(path),
	}
}

func (st *Store) Add(k string, v interface{}) error {
	doc, err := encode(k, v)
	if err != nil {
		return err
	}
	st.Lock()
	st.index.Add([]byte(k), doc)
	st.Unlock()
	return nil
}

func (st *Store) Set(k string, v interface{}) error {
	doc, err := encode(k, v)
	if err != nil {
		return err
	}
	st.Lock()
	st.index.Set([]byte(k), doc)
	st.Unlock()
	return nil
}

func (st *Store) Get(k string, ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	if doc := st.index.GetDoc([]byte(k)); doc != nil {
		if err := decode(doc, ptr); err != nil {
			return err
		}
		return nil
	}
	return ErrNotFound
}

func (st *Store) Del(k string) {
	st.Lock()
	st.index.Del([]byte(k))
	st.Unlock()
}

func (st *Store) All(ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	docs := bytes.Join(st.index.All(), []byte{','})
	docs = append([]byte{'['}, append(docs, byte(']'))...)
	if err := decode(docs, ptr); err != nil {
		return err
	}
	return nil
}

func (st *Store) Match(qry *regexp.Regexp, ptr interface{}) error {
	st.RLock()
	defer st.RUnlock()
	docs := bytes.Join(st.index.Match(qry), []byte{','})
	docs = append([]byte{'['}, append(docs, byte(']'))...)
	if err := decode(docs, ptr); err != nil {
		return err
	}
	return nil
}

func (st *Store) Query(ptr interface{}, qry ...string) error {
	st.RLock()
	defer st.RUnlock()
	docs := bytes.Join(st.index.Query(qry...), []byte{','})
	docs = append([]byte{'['}, append(docs, byte(']'))...)
	if err := decode(docs, ptr); err != nil {
		return err
	}
	return nil
}

func (st *Store) TestQuery(ptr interface{}, qry [][]byte) error {
	st.RLock()
	defer st.RUnlock()
	docs := bytes.Join(st.index.TestQuery(qry), []byte{','})
	docs = append([]byte{'['}, append(docs, byte(']'))...)
	if err := decode(docs, ptr); err != nil {
		return err
	}
	return nil
}

func (st *Store) TestQueryOne(ptr interface{}, qry [][]byte) bool {
	st.RLock()
	defer st.RUnlock()
	docs := st.index.TestQuery(qry)
	if len(docs) != 1 {
		return false
	}
	if err := decode(docs[0], ptr); err != nil {
		return false
	}
	return true
}
