package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cagnosolutions/idx"
)

var t *idx.Tree = idx.NewTree()

func main() {
	http.HandleFunc("/btree/set", HandleCORS(set))
	http.HandleFunc("/btree/del", HandleCORS(del))
	http.HandleFunc("/btree/get", HandleCORS(get))
	http.HandleFunc("/btree/clr", HandleCORS(clr))
	http.HandleFunc("/btree/mlt", HandleCORS(mlt))
	http.ListenAndServe(":8080", nil)
}

func HandleCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fn(w, r)
	}
}

func set(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		if k := r.FormValue("key"); k != "" {
			i, _ := strconv.Atoi(k)
			t.Set([]byte(k), i)
			fmt.Fprintf(w, `%s`, t)
			return
		}
	}
	fmt.Fprintf(w, `%s`, "[]")
	return
}

func del(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		if k := r.FormValue("key"); k != "" {
			t.Del([]byte(k))
			fmt.Fprintf(w, `%s`, t)
			return
		}
	}
	fmt.Fprintf(w, `%s`, "[]")
	return
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		if k := r.FormValue("key"); k != "" {
			r := t.Get([]byte(k))
			if r == nil {
				fmt.Fprintf(w, `%s`, "ERROR")
				return
			}
			fmt.Fprintf(w, `%v`, r.Val)
			return
		}
	}
	fmt.Fprintf(w, `%s`, "[]")
	return
}

func clr(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == "POST" {
		t.Close()
		t = idx.NewTree()
		fmt.Fprintf(w, `%s`, t)
		return
	}
	fmt.Fprintf(w, `%s`, "[]")
	return
}

func mlt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	w.WriteHeader(http.StatusOK)
	num, err := strconv.Atoi(r.FormValue("key"))
	if err != nil {
		fmt.Fprintf(w, "[]")
		return
	}
	if r.Method == "POST" {
		for i := 1; i <= num; i++ {
			k := fmt.Sprintf("%0.3d", i)
			t.Set([]byte(k), i)
		}
		fmt.Fprintf(w, `%s`, t)
		return
	}
	fmt.Fprintf(w, `%s`, "[]")
	return
}
