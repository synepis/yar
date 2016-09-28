package yar

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
)

type MyContext interface {
	Value(key interface{}) interface{}
}

type emptyCtx struct{}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}

// A valueCtx carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Context.
type valueCtx struct {
	MyContext
	key, val interface{}
}

func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.MyContext.Value(key)
}

// The provided key must be comparable.
func WithValue(parent MyContext, key, val interface{}) MyContext {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}

type MyRequest struct {
	ctx MyContext
}

func (r *MyRequest) Context() MyContext {
	if r.ctx != nil {
		return r.ctx
	}
	return &emptyCtx{}
}

func (r *MyRequest) WithContext(ctx MyContext) *MyRequest {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(MyRequest)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

type HHandler func(w http.ResponseWriter, r *http.Request, ps Params)

type RequestContextKey int

const ROUTE_PARAMS_KEY RequestContextKey = 0

type Route struct {
	Path *Path
	// Handlers map[string]http.Handler // Method handlers
	Handlers map[string]HHandler // Method handlers
}

func NewRoute(urlPattern string) *Route {
	return &Route{
		Path: NewPath(urlPattern),
		// Handlers: make(map[string]http.Handler),
		Handlers: make(map[string]HHandler),
	}
}

type Router struct {
	routeTrie               Tree
	notFoundHandler         http.Handler
	methodNotAllowedHandler http.Handler
	//TODO: optionsHandle          Handle
}

func New() *Router {
	return &Router{
		routeTrie: Tree(NewRouteTrie()),
		//routeTrie: Tree(NewSegmentTree()),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params := r.routeTrie.FindRoute(req.URL.Path)
	reqWithParams := req.WithContext(context.WithValue(req.Context(), ROUTE_PARAMS_KEY, params))
	//appendParameters(req, params)
	if route != nil { // Found route
		handler := route.Handlers[req.Method]
		if handler != nil { // Found method handler
			handler(w, reqWithParams, params)
		} else {
			r.handleMethodNotAllowed(w, req)
		}
	} else {
		r.handleNotFound(w, req)
	}
}

func appendParameters(r *http.Request, params Params) {
	totalLen := len(r.URL.RawQuery)
	for _, p := range params {
		totalLen += len(p.Key) + len(p.Value) + 2
	}
	if totalLen > 0 {
		totalLen--
	}
	query := make([]byte, totalLen)
	pos := 0
	for pos < len(r.URL.RawQuery) {
		query[pos] = r.URL.RawQuery[pos]
		pos++
	}
	for pi, p := range params {
		for i := 0; i < len(p.Key); i++ {
			query[pos] = p.Key[i]
			pos++
		}
		query[pos] = '='
		pos++
		for i := 0; i < len(p.Value); i++ {
			query[pos] = p.Value[i]
			pos++
		}
		if pi != len(params)-1 {
			query[pos] = '&'
			pos++
		}
	}
	r.URL.RawQuery = string(query)
}

func (r *Router) handleMethodNotAllowed(w http.ResponseWriter, req *http.Request) {
	if r.methodNotAllowedHandler != nil {
		r.methodNotAllowedHandler.ServeHTTP(w, req)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	if r.notFoundHandler != nil {
		r.notFoundHandler.ServeHTTP(w, req)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (r *Router) AddHandler(method, path string, handler HHandler) {
	route, _ := r.routeTrie.FindRoute(path)
	// If route doesn't exist, first create it
	if route == nil {
		route = NewRoute(path)
		r.routeTrie.AddRoute(route)
	}
	// Add method handler
	if route.Handlers[method] != nil {
		panic(fmt.Sprintf("cannot register the same path ('%s') and method ('%s') more than once", path, method))
	}
	route.Handlers[method] = handler
}

// func (r *Router) AddHandleFunc(method, path string, handlerFunc http.HandlerFunc) {
// 	r.AddHandler(method, path, handlerFunc)
// }

func (r *Router) AddHandle(method, path string, handler HHandler) {
	r.AddHandler(method, path, handler)
}

func (r *Router) Get(path string, handlerFunc HHandler) {
	r.AddHandle("GET", path, handlerFunc)
}

func (r *Router) Post(path string, handlerFunc HHandler) {
	r.AddHandle("POST", path, handlerFunc)
}

func (r *Router) Put(path string, handlerFunc HHandler) {
	r.AddHandle("PUT", path, handlerFunc)
}

func (r *Router) Patch(path string, handlerFunc HHandler) {
	r.AddHandle("PATCH", path, handlerFunc)
}

func (r *Router) Delete(path string, handlerFunc HHandler) {
	r.AddHandle("DELETE", path, handlerFunc)
}

func GetParam(r *http.Request, key string) string {
	params := r.Context().Value(ROUTE_PARAMS_KEY).(Params)
	return params.Value(key)
	// if params == nil {
	// 	return "", errors.New("no parameters were intialized for this reqest")
	// }
	// paramValue, ok := params[paramName]
	// if !ok {
	// 	return "", errors.New("no such parameter was found")
	// }
	// return paramValue, nil
}
