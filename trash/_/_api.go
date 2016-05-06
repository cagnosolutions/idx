package adb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

/*
 *                              -- REpresentational State Transfer, REST "Spec" --
 *
 *	                      REST was defined by Roy Thomas Fielding in his 2000 PhD dissertation
 *	                  "Architectural Styles and the Design of Network-based Software Architectures".
 *
 *	+========================================+================+================+=================+===================+
 *	|  Uniform Resource Identifier, aka URI  |  "GET" Method  |  "PUT" Method  |  "POST" Method  |  "DELETE" Medhod  |
 *	+========================================+================+================+=================+===================+
 *	|                                        |  List the URIs |   Replace the  |   Create a new  |                   |
 *	|                                        |  and any other |     entire     |   entry in the  | Delete the entire |
 *	|	http://example.io/api/resource		 |  optional de-  |   collection   |  collection and | contents of this  |
 *	|										 |  tails of the  |  with another  |  return an auto |    collection     |
 *	|										 |   collection   |   collection   |   assigned ID   |                   |
 *	+----------------------------------------+----------------+----------------+-----------------+-------------------+
 *	|                                        |    Retrieve a  | Replace, or if |                 |                   |
 *	|                                        | representation |  it does not   |   *This is not  | Delete the chosen |
 *	|  http://example.io/api/resource/item9  |  of the chosen | exist, create  |  typically used |   member of this  |
 *	|                                        |  member of the | chosen member  |   for anything  |     collection    |
 *	|                                        |   collection   | in collection  |                 |                   |
 *	+----------------------------------------+----------------+----------------+-----------------+-------------------+
 *
 */

func encJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		// http.StatusInternalServerError, 500
		fmt.Fprintf(w, `{"error":true,"status":%q,"code":%d}`, http.StatusText(500), 500)
	}
	return
}

func decJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (db *DB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ignore options and favicon requests...
	if r.Method == "OPTIONS" || r.URL.Path == "/favicon.ico" {
		return
	}
	// check the referer header...
	if !strings.Contains(r.Referer(), r.Host) {
		// http.StatusBadRequest, 400
		fmt.Fprintf(w, "<h1>%s</h1></p>Status %d</p>", http.StatusText(400), 400)
		return
	}
	// check for http basic authentication...
	if user, pass, ok := r.BasicAuth(); ok {
		if db.checkBasicAuth(user, pass) {
			// http.StatusUnauthorized, 401
			fmt.Fprintf(w, "<h1>%s</h1></p>Status %d</p>", http.StatusText(401), 401)
			return
		}
	}
	// parse url request path
	path := strings.Split(r.URL.Path, "/")[:2]
	// match method case up with data handler
	switch r.Method {
	case "GET":
		if len(path) == 1 {
			stats := struct {
				Store string `json:"store"`
				Stats string `json:"stats"`
			}{
				Store: path[0],
				Stats: "This part of the api has not been implemented yet",
			}
			encJSON(w, stats)
			return
		}
		if len(path) == 2 {
			var d json.RawMessage
			db.Get(path[0], path[1], &d)
			fmt.Fprintf(w, "%s", d)
			return
		}
		return
	case "POST":
		return
	case "PUT":
		return
	case "DELETE":
		return
	case "HEAD":
		return
	}
	return
}

func (db *DB) checkBasicAuth(user, pass string) bool {
	db.RLock()
	defer db.RUnlock()
	st, ok := db.namespace("_users")
	if !ok {
		return false
	}
	var v map[string]interface{}
	if err := st.Get(user, &v); err != nil {
		return false
	}
	return v["pass"].(string) == pass
}
