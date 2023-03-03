package algnhsa_test

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/akrylysov/algnhsa"
)

func addHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := strconv.Atoi(r.FormValue("first"))
	s, _ := strconv.Atoi(r.FormValue("second"))
	w.Header().Set("X-Hi", "foo")
	fmt.Fprintf(w, "%d", f+s)
}

func contextHandler(w http.ResponseWriter, r *http.Request) {
	lambdaEvent, ok := algnhsa.APIGatewayV2RequestFromContext(r.Context())
	if ok {
		fmt.Fprint(w, lambdaEvent.RequestContext.AccountID)
	}
}

func main() {
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/context", contextHandler)
	algnhsa.ListenAndServe(http.DefaultServeMux, nil)
}
