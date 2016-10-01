package yar

import "strings"

type node struct {
	char      byte
	route     *Route // Only leaf nodes have a route != nil
	paramKey  string
	maxParams int // Maximum number of params that would need to be allocated for any path in this node's subtree
	parent    *node
	children  []*node
}

func (n *node) AddChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) GetChild(b byte) *node {
	for _, c := range n.children {
		if c.char == b {
			return c
		}
	}
	return nil
}

type routeTrie struct {
	root node
}

func newRouteTrie() *routeTrie {
	return &routeTrie{}
}

func (rt *routeTrie) AddRoute(route *Route) {
	current := &rt.root
	pattern := route.Path.UrlPattern
	paramCnt := 0
	maxParams := max(current.maxParams, len(route.Path.ParamKeys))
	for i := 0; i < len(pattern); i++ {
		next := current.GetChild(pattern[i])
		// Extract parameter if it exists, will be empty otherwise
		char := pattern[i]
		paramKey := ""
		if isParameter(pattern[i]) {
			char = pattern[i]
			paramKey = prefixUntilSlash(pattern[i+1:])
			i += len(paramKey) // Advance to next path part
			paramCnt++
		}
		// If no next node exists create one
		if next == nil {
			mustNotCollide(current, char, paramKey)
			next = &node{
				char:     char,
				parent:   current,
				paramKey: paramKey,
			}
			current.AddChild(next)
		}
		// Add route if this is a leaf node
		if i == len(pattern)-1 {
			mustBeUniquePath(next)
			next.route = route
			next.maxParams = maxParams
		}
		current.maxParams = maxParams
		current = next
	}
}

// Ensuring there is no path collision
func mustNotCollide(node *node, char byte, paramKey string) {
	if isParameter(char) {
		if (node.GetChild(':') != nil && char == '*') ||
			(node.GetChild('*') != nil && char == ':') {
			panic("parameter and wilcard types cannot be in the same path part, e.g.:[/user/:user_id,/user/*user_id]")
		} else if node.GetChild(char) != nil && node.GetChild(char).paramKey != paramKey {
			panic("cannot have two different parameter names for the same path part, e.g.: [/user/:user_id,/user/:user]")
		} else if len(node.children) > 0 {
			panic("parameter and static parts of the path cannot be in the same place, e.g.: [/blog/:blog_id,/blog/new]")
		}
	} else if node.GetChild(':') != nil || node.GetChild('*') != nil {
		panic("parameter and static parts of the path cannot be in the same place, e.g.: [/blog/:blog_id,/blog/new]")

	}
}

func mustBeUniquePath(n *node) {
	if n.route != nil {
		panic("cannot insert the same path twice")
	}
}

func isParameter(char byte) bool {
	return char == ':' || char == '*'
}

func (rt *routeTrie) FindRoute(path string) (*Route, Params) {
	var params Params
	paramCnt := 0
	current := &rt.root
	for i := 0; i < len(path); i++ {
		var next *node
		// Static part
		if !isParameter(path[i]) {
			next = current.GetChild(path[i])
		}
		// If there is no static part, check for parameters
		if next == nil {
			if current.GetChild(':') != nil {
				next = current.GetChild(':')
				paramVal := prefixUntilSlash(path[i:])
				if params == nil { // Lazy init
					params = make(Params, next.maxParams)
				}
				params[paramCnt] = Param{Key: next.paramKey, Value: paramVal}
				paramCnt++
				i += len(paramVal) - 1 // Advance to next path part
			} else if current.GetChild('*') != nil {
				next = current.GetChild('*')
				paramVal := path[i:]
				if params == nil { // Lazy init
					params = make(Params, next.maxParams)
				}
				params[paramCnt] = Param{Key: next.paramKey, Value: paramVal}
				paramCnt++
				return current.GetChild('*').route, params[:paramCnt] // If wildcard we return immediately
			}
		}
		// Unrecognized path
		if next == nil {
			return nil, nil
		}
		current = next
	}
	if current != nil && current.route != nil { // Found a leaf node
		return current.route, params[:paramCnt]
	}
	return nil, nil // Unrecognized path
}

func prefixUntilSlash(str string) string {
	index := strings.Index(str, "/")
	if index > 0 {
		return str[:index]
	}
	return str
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
