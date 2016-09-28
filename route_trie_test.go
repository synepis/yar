package yar

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RouteTestCase struct {
	Path   string
	Params map[string]string
}

func TestFindingRoutes(t *testing.T) {
	rt := NewRouteTrie()

	rt.AddRoute(NewRoute("/"))
	rt.AddRoute(NewRoute("/user"))
	rt.AddRoute(NewRoute("/user/"))
	rt.AddRoute(NewRoute("/user/:user_id/contact/:contact_id"))
	rt.AddRoute(NewRoute("/user/:user_id"))
	rt.AddRoute(NewRoute("/blog/:blog_id"))
	rt.AddRoute(NewRoute("/blog/:blog_id/new"))
	rt.AddRoute(NewRoute("/blog/:blog_id/edit"))
	rt.AddRoute(NewRoute("/blog/:blog_id/post/:post_id"))
	rt.AddRoute(NewRoute("/static/*filepath"))
	rt.AddRoute(NewRoute("/images/:user_id/static/*filepath"))
	// rt.AddRoute(NewRoute("/not_static/*filepath/:dummy_var")) // Should fail adding this route!
	rt.AddRoute(NewRoute("/unicode日本語/:⌘"))

	var testCases = []RouteTestCase{
		RouteTestCase{"/", make(map[string]string)},
		RouteTestCase{"/user", make(map[string]string)},
		RouteTestCase{"/user/", make(map[string]string)},
		RouteTestCase{"/blog/1", map[string]string{"blog_id": "1"}},
		RouteTestCase{"/blog/2/post/3", map[string]string{"blog_id": "2", "post_id": "3"}},
		RouteTestCase{"/blog/4/new", map[string]string{"blog_id": "4"}},
		RouteTestCase{"/blog/5/edit", map[string]string{"blog_id": "5"}},
		RouteTestCase{"/static/a/b/c/test.gif", map[string]string{"filepath": "a/b/c/test.gif"}},
		RouteTestCase{"/images/6/static/a/b/c/test.gif", map[string]string{"user_id": "6", "filepath": "a/b/c/test.gif"}},
		RouteTestCase{"/unicode日本語/unicodePatram日本語", make(map[string]string)},
		RouteTestCase{"/blog/:blog_id", map[string]string{"blog_id": ":blog_id"}},
	}

	for _, tc := range testCases {
		r, params := rt.FindRoute(tc.Path)
		assert.NotNil(t, r)
		for p := range tc.Params {
			assert.Equal(t, tc.Params[p], params.Value(p))
		}
	}
}

func PrintTree(n *Node, depth int) {
	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("%s\n", []byte{n.char})
	for _, c := range n.children {
		if c != nil {
			PrintTree(c, depth+1)
		}
	}
}

// func TestAddingDuplicatePathsCausesPanic(t *testing.T) {
// 	panicked := false

// 	func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				panicked = true
// 			}
// 		}()

// 		rt := NewRouteTrie()
// 		rt.AddRoute(NewRoute("/user"))
// 		rt.AddRoute(NewRoute("/user"))
// 	}()

// 	assert.True(t, panicked)
// }

// func TestAddingPathsWithParameterCollisionCausesPanic(t *testing.T) {
// 	panicked := make([]bool, 4)

// 	testPanic := func(idx int, path1, path2 string) {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				panicked[idx] = true
// 			}
// 		}()

// 		panicked[idx] = false
// 		rt := NewRouteTrie()
// 		rt.AddRoute(NewRoute(path1))
// 		rt.AddRoute(NewRoute(path2))
// 	}

// 	testPanic(0, "/user/:user_id", "/user/*user_id]")
// 	testPanic(1, "/user/:user_id", "/user/:user]")
// 	testPanic(2, "/user/:user_id", "/user/new]")
// 	testPanic(3, "/user/new", "/user/*user_id]")

// 	for i := range panicked {
// 		assert.True(t, panicked[i])
// 	}
// }
