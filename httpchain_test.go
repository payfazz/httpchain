package httpchain_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/payfazz/httpchain"
)

func genMiddleware(id string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Before", id)
			next(w, r)
			w.Header().Add("After", id)
		}
	}
}

var testData = "testdata"

func check(t *testing.T, handler interface{}) {
	h := httpchain.Chain(
		genMiddleware("1"),
		nil,
		[]interface{}{
			genMiddleware("2"),
			genMiddleware("3"),
			[]interface{}{
				genMiddleware("4"),
			},
			genMiddleware("5"),
		},
		genMiddleware("6"),
		handler,
		func(http.HandlerFunc) http.HandlerFunc {
			panic("should not go here")
		},
	)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h(res, req)

	if res.Body.String() != testData {
		t.Errorf("invalid res body")
	}
	if !reflect.DeepEqual(res.Header()["Before"], []string{"1", "2", "3", "4", "5", "6"}) {
		t.Errorf("invalid res header: Before")
	}
	if !reflect.DeepEqual(res.Header()["After"], []string{"6", "5", "4", "3", "2", "1"}) {
		t.Errorf("invalid res header: After")
	}
}

func testHandler(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, testData) }

func TestChain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", testHandler)

	// for http.HandlerFunc
	check(t, testHandler)

	// for http.Handler
	check(t, mux)

	// for func(*http.Request) http.HandlerFunc
	check(t, func(*http.Request) http.HandlerFunc { return testHandler })
}

func TestInvalidChain(t *testing.T) {
	// this test is just for code coverage

	gotPanic := false

	func() {
		defer func() { gotPanic = recover() != nil }()
		httpchain.Chain("invalid middleware")
	}()

	if !gotPanic {
		t.Errorf("should panic")
	}
}

func TestDoc(t *testing.T) {
	// this test is from documentation of Chain function
	var h http.HandlerFunc
	var m func(http.HandlerFunc) http.HandlerFunc
	var ms [2]func(http.HandlerFunc) http.HandlerFunc

	h = testHandler
	m = genMiddleware("1")
	ms = [2]func(http.HandlerFunc) http.HandlerFunc{genMiddleware("2"), genMiddleware("3")}

	req1 := httptest.NewRequest("GET", "/", nil)
	res1 := httptest.NewRecorder()
	all1 := m(ms[0](ms[1](h)))
	all1(res1, req1)

	req2 := httptest.NewRequest("GET", "/", nil)
	res2 := httptest.NewRecorder()
	all2 := httpchain.Chain(m, ms, h)
	all2(res2, req2)

	if !reflect.DeepEqual(res1, res2) {
		t.Errorf("Chain should have same behaviour as manual chaining by hand")
	}
}

func TestHandlerInMiddle(t *testing.T) {
	gotPanic := false

	identityMiddleware := func(next http.HandlerFunc) http.HandlerFunc { return next }
	panicMiddleware := func(http.HandlerFunc) http.HandlerFunc { panic("should not go here") }
	handler := func(http.ResponseWriter, *http.Request) {}
	func() {
		defer func() { gotPanic = recover() != nil }()
		httpchain.Chain(
			identityMiddleware,
			[]interface{}{
				identityMiddleware,
				[]interface{}{
					identityMiddleware,
					handler,
					panicMiddleware,
				},
				panicMiddleware,
			},
			panicMiddleware,
		)
	}()

	if gotPanic {
		t.Errorf("should not panic")
	}
}
