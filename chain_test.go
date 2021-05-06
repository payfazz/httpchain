package httpchain_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/payfazz/httpchain"
)

func ExampleChain() {
	middleware1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Before", "m1")
			next(w, r)
			w.Header().Add("After", "m1")
		}
	}

	middleware2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Before", "m2")
			next(w, r)
			w.Header().Add("After", "m2")
		}
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "x")
	}

	all := httpchain.Chain(
		middleware1,
		middleware2,
		handler,
	)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	all(res, req)

	fmt.Printf(
		"%s,%s,%s,%s,%s\n",
		res.Header().Values("Before")[0],
		res.Header().Values("Before")[1],
		res.Body.String(),
		res.Header().Values("After")[0],
		res.Header().Values("After")[1],
	)

	// Output: m1,m2,x,m2,m1
}
