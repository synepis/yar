package yar

import (
	"context"
	"fmt"
	"net/http"
)

type RequestContextKey int

const ROUTE_PARAMS_KEY RequestContextKey = 0

type Route struct {
	Path     *Path
	Handlers map[string]http.Handler // Method handlers
}

func NewRoute(urlPattern string) *Route {
	return &Route{
		Path:     NewPath(urlPattern),
		Handlers: make(map[string]http.Handler),
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
	if route != nil { // Found route
		handler := route.Handlers[req.Method]
		if handler != nil { // Found method handler
			handler.ServeHTTP(w, reqWithParams)
		} else {
			r.handleMethodNotAllowed(w, reqWithParams)
		}
	} else {
		r.handleNotFound(w, reqWithParams)
	}
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

func (r *Router) AddHandler(method, path string, handler http.Handler) {
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

func (r *Router) AddHandleFunc(method, path string, handlerFunc http.HandlerFunc) {
	r.AddHandler(method, path, handlerFunc)
}

func (r *Router) AddHandle(method, path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandler(method, path, http.HandlerFunc(handlerFunc))
}

func (r *Router) Get(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandle("GET", path, handlerFunc)
}

func (r *Router) Post(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandle("POST", path, handlerFunc)
}

func (r *Router) Put(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandle("PUT", path, handlerFunc)
}

func (r *Router) Patch(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandle("PATCH", path, handlerFunc)
}

func (r *Router) Delete(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
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
