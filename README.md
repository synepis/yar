# YAR ![Builds status](https://travis-ci.org/synepis/yar.svg?branch=master "Optional Title") [![Coverage Status](https://coveralls.io/repos/github/synepis/yar/badge.svg?branch=master)](https://coveralls.io/github/synepis/yar?branch=master)
Golang HTTP Router - **Y**et **A**nother Http **R**outer

*Note:* This is a **test project** made for learning purposes, not meant to be used for any kind of production.

Why another Go HTTP router? One of the problems of writing a HTTP router is how to pass the path parameters to the method handlers. So far I've seen 3 approaches
- *Custom context:* Most of the routers out there utilize a custom context object to pass path parameters and thus locking you into their implementation. However, this is the fastest option.
- *Global map*: This is another popular option which I believe to be suboptimal when used with a lot of concurrent requests. Some side effects include the fact that multiple routers share the same global map, hard to control the lifecycle of the map, need to worry about clearing out the request's context from the map once the request has been served.
- *URL query rewriting* - This option writes the found parameters to the querystring of the request. This is quite slow once you have to parse the URL again to get the parameters. 

YAR approach uses Go's 1.7 http.Request.Context to pass path parameters. This way there's no locking and no opting into custom handler implementation. This is not the fastest approach, but it's not too slow and we don't need to change/write to the request which could be accessed concurrently.

## Design:
- Prefix Trie used to find routes
- Attaching parameters to http.Request.Context
- Has native NotFound, MethodNotAllowed and OPTIONS handlers (you can use your own if you prefer)
- Each path pattern can be matched by only one route (if any collision are possible the router will panic; thus letting you know immediately rather than later on while running the application)

*Things missing and on the TODO list:*
- Trailing slash ignoring - if you wish to have '/user' anb '/user/' point to the same handler you have to add both paths
- Case Sensitivity - router is currently case sensitive, plan is to add option to ignore case
- Panic recovery - I believe this should be handled by a middleware

## Usage:
Before using YAR, get it by using go get:
```
go get github.com/synepis/yar
```

To start using it:
```go
router := yar.NewRouter()
router.ShouldLog = true // By default is true to help with debugging,
                   // set to false for production use
router.ShouldHandleOptions = true // Let YAR automatically respond with allowed methods for a resource

// Route registrations here

http.ListenAndServe(":8080", router)
```

### Registering routes:
You can register any route using either a http.Handler,http.HandlerFunc or simply any function which has the 'func(http.ResponseWriter, *http.Request); signature. Beside those there are a few predefined methods you can use.

### Parameters
#### Regular parameter
A regular will match any text inbetween two '/' symbols (a path segment).
To add a regular parameter to you path pattern, just prefix it with a ':' symbol:
```go
router.Get("/user/:user_id/details", func(w http.ResponseWriter, r *http.Request) {})
```

#### Wildcard
A wildcard parameter must be placed at the end of a path. It will match all text after it.
To add a wildcard parameter to you path pattern, just prefix it with a '*' symbol:
```go
router.Get("/static/*filepath", func(w http.ResponseWriter, r *http.Request) {})
```
#### To read parameters:
```go
user := yar.GetParam(r, "user") // r is *http.Request

params := yar.GetParams(r)
for i := range params {
    w.Write([]byte(fmt.Sprintf("(%s, %s) ", params[i].Key, params[i].Value)))
}
```

### Custom handlers:
To se your own NotFound or MethodNotAllowed handlers:
```go
router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom NotFound\n"))
})

router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom MethodNotAllowed\n"))
})
```



### Example:
```go
package main

import (
	"fmt"
	"net/http"

	"github.com/synepis/yar"
)

func main() {
	router := yar.NewRouter()
	router.ShouldHandleOptions = true
	router.ShouldLog = true // On by default, but here set explicitly

	// curl localhost:8080
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
	// curl localhost:8080/hello/gordon
	router.Get("/hello/:user", func(w http.ResponseWriter, r *http.Request) {
		user := yar.GetParam(r, "user")
		w.Write([]byte("Hello " + user + "\n"))
	})
	// curl localhost:8080/static/images/thumbnails/thumb1.png
	router.Get("/static/*filepath", func(w http.ResponseWriter, r *http.Request) {
		filepath := yar.GetParam(r, "filepath")
		w.Write([]byte("Serving: /static/" + filepath + "\n"))
	})

	// curl localhost:8080/user/gordon/files/documents/doc1.txt
	router.Get("/user/:user/files/*filepath", func(w http.ResponseWriter, r *http.Request) {
		user := yar.GetParam(r, "user")
		filepath := yar.GetParam(r, "filepath")

		resp := fmt.Sprintf("Serving: /user/%s/files/%s\n", user, filepath)
		w.Write([]byte(resp))
	})

	// curl localhost:8080/not-found
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom NotFound\n"))
	})

	// curl -X POST localhost:8080/hello/gordon
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom MethodNotAllowed\n"))
	})

	// curl -X OPTIONS localhost:8080/options-example
	router.Get("/options-example", func(w http.ResponseWriter, r *http.Request) {})
	router.Post("/options-example", func(w http.ResponseWriter, r *http.Request) {})
	router.Patch("/options-example", func(w http.ResponseWriter, r *http.Request) {})
	router.Delete("/options-example", func(w http.ResponseWriter, r *http.Request) {})

	// YAR only parses path parameters, not the query
	// curl localhost:8080/blog/123/post/456?q1=111&q2=222
	router.Get("/blog/:blog_id/post/:post_id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Path params: "))
		params := yar.GetParams(r)
		for i := range params {
			w.Write([]byte(fmt.Sprintf("(%s, %s) ", params[i].Key, params[i].Value)))
		}
		w.Write([]byte("\n"))

		w.Write([]byte("Query params: "))
		for param := range r.URL.Query() {
			w.Write([]byte(fmt.Sprintf("(%s, %s) ", param, r.URL.Query().Get(param))))
		}
		w.Write([]byte("\n"))
	})

	http.ListenAndServe(":8080", router)
}
```
#### Output of curl commands in above example:
```
2016/10/01 14:03:54 [YAR] [GET] [/] -> [Found: /]
2016/10/01 14:03:59 [YAR] [GET] [/hello/gordon] -> [Found: /hello/:user]
2016/10/01 14:04:02 [YAR] [GET] [/static/images/thumbnails/thumb1.png] -> [Found: /static/*filepath]
2016/10/01 14:04:05 [YAR] [GET] [/user/gordon/files/documents/doc1.txt] -> [Found: /user/:user/files/*filepath]
2016/10/01 14:04:09 [YAR] [GET] [/not-found] -> [Not Found]
2016/10/01 14:04:11 [YAR] [POST] [/hello/gordon] -> [Method Not Allowed]
2016/10/01 14:04:14 [YAR] [OPTIONS] [/options-example] -> [Handling OPTIONS]
2016/10/01 14:04:17 [YAR] [GET] [/blog/123/post/456?q1=111] -> [Found: /blog/:blog_id/post/:post_id]
```

## Performance
The router has decent performance, even though it is implemented with the simpler trie rather than a full-on radix tree. However the biggest impact on performance is the usage of context.Context and http.Request.Context. Without it this router would be a lot closer to the fastest implementation I know of: [HttpRouter](https://github.com/julienschmidt/httprouter) which simply returns a list of parameters through a custom http handler.

Here is a benchmark of the internal RouteTrie function finding the routes (and extracting the parameters) vs. the ServeHTTP which simply calls RoutTrie's FindMethod and saves the returned parameters to the request's context.
```
BenchmarkRouteTrieTestStaticPath-8      20000000                92.5 ns/op             0 B/op          0 allocs/op
Benchmark_RouteTrie_1_Params-8          10000000               141 ns/op              32 B/op          1 allocs/op
Benchmark_RouteTrie_5_Params-8           2000000               613 ns/op             160 B/op          1 allocs/op
Benchmark_RouteTrie_10_Params-8          1000000              1190 ns/op             320 B/op          1 allocs/op
Benchmark_RouteTrie_20_Params-8           500000              2666 ns/op             640 B/op          1 allocs/op

Benchmark_Router_StaticPath-8           20000000               113 ns/op               0 B/op          0 allocs/op
Benchmark_Router_1_Params-8              1000000              1398 ns/op             496 B/op          6 allocs/op
Benchmark_Router_5_Params-8              1000000              1328 ns/op             496 B/op          6 allocs/op
Benchmark_Router_10_Params-8             1000000              2211 ns/op             656 B/op          6 allocs/op
Benchmark_Router_20_Params-8              300000              4406 ns/op             976 B/op          6 allocs/op
```

### HttpRouter's benchmarks
[HttpRouter](https://github.com/julienschmidt/httprouter) has a great set of [benchmarks](https://github.com/julienschmidt/go-http-routing-benchmark) which I've adapted to use YAR and ran locally.

The forked repo of HttpRouter tests is [here](https://github.com/synepis/go-http-routing-benchmark)

*Note: The Gin router seems to be the fastest, it uses HttpRouter underneath but with some additional optimizations.*

Here are the results:



#### Benchmark results analysis summary

In order to compare the routers' performance, each of the routers has been run against
the full set of benchmarks. The number of operations per second has been chosen
as the key performance indicator. For each of the benchmarks the top `ops/s` result
has been denoted as 100% result, with the rest compared relatively to it.

![benchmark matrix](benchmark_static/router_x_benchmark.png)

For example: for the `GithubStatic` benchmark `Denco` obtained the best result of
149253 operations per second. For this benchmark Yar performed at 8904 operations
per second, giving it a score of 53% (of the best result).

All of the benchmarks were considered equally significant and the average for each
of the routers got calculated:

![ benchmark summary](benchmark_static/summary.png)

The full calculations can be found in the [Excel spreadsheet](benchmark_static/benchmar_results_analysis.xlsx).


## Credits
- The project was inspired and uses some of the benchmark code from [HttpRouter](https://github.com/julienschmidt/httprouter) 
- Special thanks to [nietaki](https://github.com/nietaki) for using his Excel prowess in creating the nice benchmark comparison overview.
