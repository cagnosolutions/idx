package adb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

const DB_PATH = "db/"

type DB struct {
	stores map[string]*Store
	sync.RWMutex
}

func NewDB() *DB {
	if err := os.MkdirAll(DB_PATH, 0755); err != nil {
		panic(err)
	}
	files, err := ioutil.ReadDir(DB_PATH)
	if err != nil {
		panic(err)
	}
	db := &DB{stores: make(map[string]*Store, 0)}
	for _, file := range files {
		if file.IsDir() {
			db.AddStore(file.Name())
		}
	}
	return db
}

func (db *DB) namespace(store string) (*Store, bool) {
	db.RLock()
	st, ok := db.stores[store]
	db.RUnlock()
	return st, ok
}

func (db *DB) AddStore(store string) {
	if _, ok := db.namespace(store); !ok {
		db.Lock()
		mkdirs(DB_PATH + store)
		db.stores[store] = NewStore(DB_PATH + store + "/" + store)
		db.Unlock()
	}
}

func (db *DB) DelStore(store string) {
	if _, ok := db.namespace(store); ok {
		db.Lock()
		delete(db.stores, store)
		db.Unlock()
	}
}

func (db *DB) Add(store, key string, val interface{}) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.Add(key, val); err != nil {
		return false
	}
	return true
}

func (db *DB) Set(store, key string, val interface{}) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.Set(key, val); err != nil {
		return false
	}
	return true
}

func (db *DB) Get(store, key string, ptr interface{}) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.Get(key, ptr); err != nil {
		return false
	}
	return true
}

func (db *DB) All(store string, ptr interface{}) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.All(ptr); err != nil {
		return false
	}
	return true
}

func (db *DB) Del(store, key string) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	st.Del(key)
	return true
}

func (db *DB) Match(store, qry string, ptr interface{}) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	regexp, err := regexp.Compile(qry)
	if err != nil {
		return false
	}
	if err := st.Match(regexp, ptr); err != nil {
		return false
	}
	return true
}

func (db *DB) Query(store string, ptr interface{}, qry ...string) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.Query(ptr, qry...); err != nil {
		return false
	}
	return true
}

func (db *DB) TestQuery(store string, ptr interface{}, qry ...[]byte) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if err := st.TestQuery(ptr, qry); err != nil {
		return false
	}
	return true
}

func (db *DB) TestQueryOne(store string, ptr interface{}, qry ...[]byte) bool {
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	if !st.TestQueryOne(ptr, qry) {
		return false
	}
	return true
}

func (db *DB) Close() {
	db.Lock()
	defer db.Unlock()
	for _, st := range db.stores {
		st.index.Close()
	}
}

// func (db *DB) Auth(store, user, pass string, ptr interface{}) bool {
// 	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
// 		Logger("Auth did not receiver pointer\n")
// 		return false
// 	}
// 	query, active := deepQry(user, pass, reflect.ValueOf(ptr).Elem())
// 	query += "|" + reverse(query)
// 	regex, err := regexp.Compile(query)
// 	if err != nil {
// 		return false
// 	}
// 	st, ok := db.namespace(store)
// 	if !ok {
// 		return false
// 	}
// 	st.RLock()
// 	t1 := time.Now().UnixNano()
// 	doc := st.index.Match(regex)
// 	t2 := time.Now().UnixNano()
// 	st.RUnlock()
// 	fmt.Printf("Match took %d nanoseconds, %.2f microseconds, %.2f milliseconds, %.2f seconds\n", (t2 - t1), float64(t2-t1)/1000, float64(t2-t1)/1000/1000, float64(t2-t1)/1000/1000/1000)
// 	if len(doc) != 1 {
// 		return false
// 	}
// 	if active != "" && !bytes.Contains(doc[0], []byte(`"`+active+`":true`)) {
// 		return false
// 	}
// 	if err := json.Unmarshal(doc[0], ptr); err != nil {
// 		return false
// 	}
// 	return true
// }

/*func (db *DB) Auth(store, user, pass string, ptr interface{}) bool {
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		Logger("Auth did not receiver pointer\n")
		return false
	}
	query := buildQry(user, pass, reflect.ValueOf(ptr).Elem())
	if len(query) < 2 {
		return false
	}
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	st.RLock()
	doc := st.index.Query(query...)
	st.RUnlock()
	if len(doc) != 1 {
		return false
	}
	if err := json.Unmarshal(doc[0], ptr); err != nil {
		return false
	}
	return true
}*/

func (db *DB) Auth(store, user, pass string, ptr interface{}) bool {
	if reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		Logger("Auth did not receiver pointer\n")
		return false
	}
	query := buildTestQry(user, pass, reflect.ValueOf(ptr).Elem())
	if len(query) < 2 {
		return false
	}
	st, ok := db.namespace(store)
	if !ok {
		return false
	}
	st.RLock()
	t1 := time.Now().UnixNano()
	doc := st.index.TestQuery(query)
	t2 := time.Now().UnixNano()
	st.RUnlock()
	fmt.Printf("Test Query took %d nanoseconds, %.2f microseconds, %.2f milliseconds, %.2f seconds\n", (t2 - t1), float64(t2-t1)/1000, float64(t2-t1)/1000/1000, float64(t2-t1)/1000/1000/1000)
	if len(doc) != 1 {
		return false
	}
	if err := json.Unmarshal(doc[0], ptr); err != nil {
		return false
	}
	return true
}

const match = `(%q:"{1}?%s"{1}?).*`

var reverse = func(s string) string {
	ss := strings.SplitAfter(s, ".*")
	for i, j := 0, len(ss)-1; i < j; i, j = i+1, j-1 {
		ss[i], ss[j] = ss[j], ss[i]
	}
	return strings.Join(ss, "")
}

func deepQry(user, pass string, dat reflect.Value) (string, string) {
	var query, active string
	for i := 0; i < dat.NumField(); i++ {
		fld := dat.Type().Field(i)
		if dat.Field(i).Kind() == reflect.Struct && fld.Anonymous {
			return deepQry(user, pass, reflect.Indirect(dat.Field(i)))
		}
		switch dat.Type().Field(i).Tag.Get("auth") {
		case "username":
			query += fmt.Sprintf(match, strings.ToLower(fld.Name), user)
		case "password":
			query += fmt.Sprintf(match, strings.ToLower(fld.Name), pass)
		case "active":
			active = strings.ToLower(fld.Name)
		}
	}
	return query, active
}

func buildQry(user, pass string, dat reflect.Value) []string {
	//var query, active string
	var query []string
	for i := 0; i < dat.NumField(); i++ {
		fld := dat.Type().Field(i)
		if dat.Field(i).Kind() == reflect.Struct && fld.Anonymous {
			query = append(query, buildQry(user, pass, reflect.Indirect(dat.Field(i)))...)
		}
		switch dat.Type().Field(i).Tag.Get("auth") {
		case "username":
			query = append(query, strings.ToLower(fld.Name)+"="+user)
		case "password":
			query = append(query, strings.ToLower(fld.Name)+"="+pass)
		case "active":
			query = append(query, strings.ToLower(fld.Name)+"=true")
		}
	}
	return query
}

func buildTestQry(user, pass string, dat reflect.Value) [][]byte {
	var query [][]byte
	for i := 0; i < dat.NumField(); i++ {
		fld := dat.Type().Field(i)
		if dat.Field(i).Kind() == reflect.Struct && fld.Anonymous {
			query = append(query, buildTestQry(user, pass, reflect.Indirect(dat.Field(i)))...)
		}
		switch dat.Type().Field(i).Tag.Get("auth") {
		case "username":
			query = append(query, Eq(strings.ToLower(fld.Name), user))
		case "password":
			query = append(query, Eq(strings.ToLower(fld.Name), pass))
		case "active":
			query = append(query, Eq(strings.ToLower(fld.Name), "true"))
		}
	}
	return query
}
