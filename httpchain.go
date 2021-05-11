// Package httpchain
//
// this package provide Chain function to chain multiple middleware into single handler function
package httpchain

import (
	"net/http"
	"reflect"
)

type middleware = func(http.HandlerFunc) http.HandlerFunc

// Chain multiple middleware into single handler function
//
// Middleware is any value that have following type
// 	func(next http.HandlerFunc) http.HandlerFunc
//
// you can pass multiple middleware, slice/array of middlewares, or combination of them
//
// this function also accept following type as handler (last function in middleware chain)
// 	http.HandlerFunc
// 	http.Handler
// 	func(*http.Request) http.HandlerFunc
// 	func(http.ResponseWriter, *http.Request) error
//
// when you have following code
// 	var h http.HandlerFunc
// 	var m func(http.HandlerFunc) http.HandlerFunc
// 	var ms [2]func(http.HandlerFunc) http.HandlerFunc
// then
// 	all := Chain(m, ms, h)
// will have same effect as
// 	all := m(ms[0](ms[1](h)))
func Chain(all ...interface{}) http.HandlerFunc {
	var f http.HandlerFunc
	ms := intoMiddlewares(all)
	for i := len(ms) - 1; i >= 0; i-- {
		f = ms[i](f)
	}
	return f
}

func intoMiddlewares(as []interface{}) []middleware {
	as = flatten(as)
	ret := make([]middleware, 0, len(as))
	for _, a := range as {
		if addAsMiddleware(&ret, a) {
			continue
		}

		if addAsLastMiddleware(&ret, a) {
			break
		}

		panic("invalid argument: can't process value with type: " + reflect.TypeOf(a).String())
	}
	return ret
}

func addAsMiddleware(ret *[]middleware, a interface{}) bool {
	var b func(http.HandlerFunc) http.HandlerFunc
	if setIfConvertible(a, &b) {
		*ret = append(*ret, b)
		return true
	}
	return false
}

func addAsLastMiddleware(ret *[]middleware, a interface{}) bool {
	if addAsHandlerFunc(ret, a) {
		return true
	}

	if addAsHandler(ret, a) {
		return true
	}

	if addAsHandlerFuncGen(ret, a) {
		return true
	}

	if addAsHandlerFuncWithErr(ret, a) {
		return true
	}

	return false
}

func addAsHandlerFunc(ret *[]middleware, a interface{}) bool {
	var b http.HandlerFunc
	if setIfConvertible(a, &b) {
		*ret = append(*ret, func(http.HandlerFunc) http.HandlerFunc {
			return b
		})
		return true
	}
	return false
}

func addAsHandler(ret *[]middleware, a interface{}) bool {
	var b http.Handler
	if setIfConvertible(a, &b) {
		*ret = append(*ret, func(http.HandlerFunc) http.HandlerFunc {
			return b.ServeHTTP
		})
		return true
	}
	return false
}

func addAsHandlerFuncGen(ret *[]middleware, a interface{}) bool {
	var b func(*http.Request) http.HandlerFunc
	if setIfConvertible(a, &b) {
		*ret = append(*ret, func(http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				b(r)(w, r)
			}
		})
		return true
	}
	return false
}

func addAsHandlerFuncWithErr(ret *[]middleware, a interface{}) bool {
	var b func(http.ResponseWriter, *http.Request) error
	if setIfConvertible(a, &b) {
		*ret = append(*ret, func(http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				b(w, r)
			}
		})
		return true
	}
	return false
}

func flatten(as []interface{}) []interface{} {
	ret := make([]interface{}, 0, len(as))
	for _, a := range as {
		if a == nil {
			continue
		}

		switch reflect.TypeOf(a).Kind() {
		case reflect.Slice, reflect.Array:
			aVal := reflect.ValueOf(a)
			bs := make([]interface{}, aVal.Len())
			for i := 0; i < aVal.Len(); i++ {
				bs[i] = aVal.Index(i).Interface()
			}
			ret = append(ret, flatten(bs)...)
		default:
			ret = append(ret, a)
		}
	}
	return ret
}

func setIfConvertible(from interface{}, toPtr interface{}) bool {
	fromVal := reflect.ValueOf(from)
	fromType := fromVal.Type()
	toVal := reflect.ValueOf(toPtr).Elem()
	toType := toVal.Type()
	if fromType.ConvertibleTo(toType) {
		toVal.Set(fromVal.Convert(toType))
		return true
	}
	return false
}
