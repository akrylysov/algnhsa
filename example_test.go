package algnhsa_test

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/akrylysov/algnhsa"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("index"))
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := strconv.Atoi(r.FormValue("first"))
	s, _ := strconv.Atoi(r.FormValue("second"))
	w.Header().Set("X-Hi", "foo")
	fmt.Fprintf(w, "%d", f+s)
}

func contextHandler(w http.ResponseWriter, r *http.Request) {
	proxyReq, ok := algnhsa.ProxyRequestFromContext(r.Context())
	if ok {
		fmt.Fprint(w, proxyReq.RequestContext.AccountID)
	}
}

func Example() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/context", contextHandler)
	algnhsa.ListenAndServe(http.DefaultServeMux, nil)
}
