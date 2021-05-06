# httpchain

[![GoDoc](https://pkg.go.dev/badge/github.com/payfazz/httpchain/v2)](https://pkg.go.dev/github.com/payfazz/httpchain/v2)

Package httpchain

this package provide `Chain` function to chain multiple middleware into single handler function

## `Chain` function
`Chain` multiple middleware into single handler function

Middleware is any value that have following type
```
func(next http.HandlerFunc) http.HandlerFunc
```

you can pass multiple `middleware`, slice/array of `middleware`, or combination of them

this function also accept following type as handler (last function in middleware chain)
```
http.HandlerFunc
http.Handler
func(*http.Request) http.HandlerFunc
```
(the last one is from [go-handler](https://pkg.go.dev/github.com/payfazz/go-handler/v2) package)

when you have following code
```
var h http.HandlerFunc
var m func(http.HandlerFunc) http.HandlerFunc
var ms [2]func(http.HandlerFunc) http.HandlerFunc
```
then
```
all := Chain(m, ms, h)
```
will have same effect as
```
all := m(ms[0](ms[1](h)))
```
