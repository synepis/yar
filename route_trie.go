package yar

// TODO: Rename ensure method to be with must prefix

import "strings"

type Node struct {
	char      byte
	route     *Route // Only leaf nodes have a route != nil
	paramKey  string
	maxParams int // Maximum number of params that would need to be allocated for any path under this node
	parent    *Node
	children  [256]*Node
	childCnt  int
}

func (n *Node) AddChild(b byte, c *Node) {
	n.children[b] = c
	n.childCnt++
}

func (n *Node) GetChild(b byte) *Node {
	return n.children[b]
}

type RouteTrie struct {
	root Node
}

func NewRouteTrie() *RouteTrie {
	return &RouteTrie{}
}

func (rt *RouteTrie) AddRoute(route *Route) {
	node := &rt.root
	pattern := route.Path.UrlPattern
	paramCnt := 0
	for i := 0; i < len(pattern); i++ {
		next := node.GetChild(pattern[i])
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
			mustNotCollide(node, char, paramKey)
			next = &Node{
				char:     char,
				parent:   node,
				paramKey: paramKey,
			}
			node.AddChild(char, next)
		}
		// Add route if this is a leaf node
		if i == len(pattern)-1 {
			mustBeUniquePath(next)
			next.route = route
			next.maxParams = paramCnt
			adjustMaxParams(next)
		}
		node = next
	}
}

// Start from leaf node and bubble up the maxParam value up to root node
func adjustMaxParams(n *Node) {
	for n.parent != nil && n.parent.maxParams < n.maxParams {
		n.parent.maxParams = n.maxParams
		n = n.parent
	}
}

// Ensuring there is no path collision
func mustNotCollide(node *Node, char byte, paramKey string) {
	if isParameter(char) {
		if (node.GetChild(':') != nil && char == '*') ||
			(node.GetChild('*') != nil && char == ':') {
			panic("parameter and wilcard types cannot be in the same path part, e.g.:[/user/:user_id,/user/*user_id]")
		} else if node.GetChild(char) != nil && node.GetChild(char).paramKey != paramKey {
			panic("cannot have two different parameter names for the same path part, e.g.: [/user/:user_id,/user/:user]")
		} else if node.childCnt > 0 {
			panic("parameter and static parts of the path cannot be in the same place, e.g.: [/blog/:blog_id,/blog/new]")
		}
	} else if node.GetChild(':') != nil || node.GetChild('*') != nil {
		panic("parameter and static parts of the path cannot be in the same place, e.g.: [/blog/:blog_id,/blog/new]")

	}
}

func mustBeUniquePath(n *Node) {
	if n.route != nil {
		panic("cannot insert the same path twice")
	}
}

func isParameter(char byte) bool {
	return char == ':' || char == '*'
}

// func (rt *RouteTrie) FindRoute(path string) (*Route, map[string]string) {
func (rt *RouteTrie) FindRoute(path string) (*Route, Params) {
	var params Params
	paramCnt := 0
	node := &rt.root
	for i := 0; i < len(path); i++ {
		var next *Node
		// Static part
		if !isParameter(path[i]) {
			next = node.GetChild(path[i])
		}
		// If there is no static part, check for parameters
		if next == nil {
			if node.GetChild(':') != nil {
				next = node.GetChild(':')
				paramVal := prefixUntilSlash(path[i:])
				if params == nil { // Lazy init
					params = make(Params, next.maxParams)
				}
				params[paramCnt] = Param{Key: next.paramKey, Value: paramVal}
				paramCnt++
				i += len(paramVal) - 1 // Advance to next path part
			} else if node.GetChild('*') != nil {
				next = node.GetChild('*')
				paramVal := path[i:]
				if params == nil { // Lazy init
					params = make(Params, next.maxParams)
				}
				params[paramCnt] = Param{Key: next.paramKey, Value: paramVal}
				paramCnt++
				return node.GetChild('*').route, params[:paramCnt] // If wildcard we return immediately
			}
		}
		// Unrecognized path
		if next == nil {
			return nil, nil
		}
		node = next
	}
	if node != nil && node.route != nil { // Found a leaf node
		return node.route, params[:paramCnt]
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
