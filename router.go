package yar

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

// Used to store parameters in http.Request.Context
type requestContextKey int

const ROUTE_PARAMS_KEY requestContextKey = 0

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
	NotFoundHandler         http.Handler // If not set the default handler is used
	MethodNotAllowedHandler http.Handler // If not set the default handler is used
	ShouldHandleOptions     bool         // Print allowed methods for a resource/route
	ShouldLog               bool         // Used to help with debugging
	routeTrie               routeTrie
}

func NewRouter() *Router {
	return &Router{
		routeTrie: *newRouteTrie(),
		ShouldLog: true,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params := r.routeTrie.FindRoute(req.URL.Path)
	reqWithParams := req
	if len(params) != 0 { // Store params to context, if any
		reqWithParams = req.WithContext(context.WithValue(req.Context(), ROUTE_PARAMS_KEY, params))
	}

	if route != nil { // Found route
		handler := route.Handlers[req.Method]
		if handler != nil { // Found method handler
			if r.ShouldLog {
				log.Printf("[YAR] [%s] [%s] -> [Found: %s]", req.Method, req.URL, route.Path.UrlPattern)
			}
			handler.ServeHTTP(w, reqWithParams)
		} else if req.Method == "OPTIONS" && r.ShouldHandleOptions {
			r.handleOptions(w, reqWithParams, route)
		} else {
			r.handleMethodNotAllowed(w, reqWithParams)
		}
	} else {
		r.handleNotFound(w, reqWithParams)

	}
}

func (r *Router) handleOptions(w http.ResponseWriter, req *http.Request, route *Route) {
	if r.ShouldLog {
		log.Printf("[YAR] [%s] [%s] -> [Handling OPTIONS]", req.Method, req.URL)
	}

	methods := make([]string, 0)
	for method := range route.Handlers {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	w.Write([]byte("Allowed: " + strings.Join(methods, ", ") + "\n"))
}

func (r *Router) handleMethodNotAllowed(w http.ResponseWriter, req *http.Request) {
	if r.ShouldLog {
		log.Printf("[YAR] [%s] [%s] -> [Method Not Allowed]", req.Method, req.URL)
	}

	if r.MethodNotAllowedHandler != nil {
		r.MethodNotAllowedHandler.ServeHTTP(w, req)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	if r.ShouldLog {
		log.Printf("[YAR] [%s] [%s] -> [Not Found]", req.Method, req.URL)
	}

	if r.NotFoundHandler != nil {
		r.NotFoundHandler.ServeHTTP(w, req)
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

func (r *Router) Head(path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.AddHandle("HEAD", path, handlerFunc)
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
	params := r.Context().Value(ROUTE_PARAMS_KEY)
	if params == nil {
		return ""
	}
	return params.(Params).Value(key)
}

func GetParams(r *http.Request) Params {
	params := r.Context().Value(ROUTE_PARAMS_KEY)
	if params == nil {
		return Params{}
	}
	return params.(Params)
}
