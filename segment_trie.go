package yar

import (
	"fmt"
	"strings"
)

type SNode struct {
	pattern   string
	route     *Route // Only leaf nodes have a route != nil
	paramKey  string
	maxParams int // Maximum number of params that would need to be allocated for any path under this node
	parent    *SNode
	children  map[string]*SNode
	childCnt  int
	hasRoute  bool
}

func (n SNode) isWildcard() bool {
	return n.pattern == "*" || n.pattern == ":"
}

type SegmentTrie struct {
	root SNode
}

func NewSegmentTree() *SegmentTrie {
	return &SegmentTrie{
		root: SNode{
			children: make(map[string]*SNode),
		},
	}
}

func (t *SegmentTrie) AddRoute(route *Route) {
	parts := splitIntoParts(route.Path.UrlPattern)
	node := &t.root
	if route.Path.UrlPattern == "/" {
		t.root.route = route
		t.root.hasRoute = true
	}
	for i := 0; i < len(parts); i++ {
		pattern, varName := getNodeNames(parts[i])
		next := node.children[pattern]
		if next == nil {
			next = &SNode{
				pattern:  pattern,
				paramKey: varName,
				hasRoute: i == len(parts)-1,
				children: make(map[string]*SNode),
				route:    route,
			}
			node.children[pattern] = next
		}
		if next.isWildcard() && next.paramKey != varName {
			panic("Variable name collision in paths!")
		}
		node = next
	}
}

func splitIntoParts(path string) []string {
	allParts := strings.Split(path, "/")
	parts := []string{}
	for _, part := range allParts {
		if part != "" {
			parts = append(parts, strings.Trim(part, " "))
		}
	}
	return parts
}

func getNodeNames(part string) (string, string) {
	pattern := part
	varName := ""
	if strings.HasPrefix(part, ":") {
		pattern = ":"
		varName = part[1:]
	}
	if strings.HasPrefix(part, "*") {
		pattern = "*"
		varName = part[1:]
	}
	return pattern, varName
}

func PrintSegTree(n *SNode, d int) {
	for i := 0; i < d; i++ {
		fmt.Printf("-")
	}
	fmt.Printf("-%s\n", n.pattern)
	for _, c := range n.children {
		PrintSegTree(c, d+1)
	}
}

func (t *SegmentTrie) FindRoute(path string) (*Route, Params) {
	if path == "/" && t.root.hasRoute {
		return t.root.route, Params{}
	}
	vars := make(Params, 10)
	pCnt := 0
	// parts := splitIntoParts(path)
	node := &t.root
	// for i := 0; i < len(parts); i++ {
	for i := 1; i < len(path); i++ {
		// Find next path
		j := i + 1
		for j < len(path) && path[j] != '/' {
			j++
		}
		part := path[i:j]
		// Static path part
		//next := node.children[parts[i]]
		next := node.children[part]
		// Variable path part
		if next == nil && node.children[":"] != nil {
			next = node.children[":"]
			vars[pCnt].Key = next.paramKey
			//vars[pCnt].Value = parts[i]
			vars[pCnt].Value = part
			pCnt++
		}
		// Wildcard path part
		if next == nil && node.children["*"] != nil {
			next = node.children["*"]
			vars[pCnt].Key = next.paramKey
			vars[pCnt].Value = path[i:] // strings.Join(parts[i:], "/")
			pCnt++
			return next.route, vars[:pCnt] // Return immediately, '*' matches the rest of the path
		}
		node = next
		if node == nil {
			return nil, nil
		}
		i = j
	}

	// fmt.Printf("%v %v\n", node, pCnt)
	if node != nil && node.hasRoute {
		//fmt.Printf("%v %v\n", vars, pCnt)
		return node.route, vars[:pCnt]
	}
	return nil, nil
}
