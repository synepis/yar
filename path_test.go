package yar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlSingleParameters(t *testing.T) {
	p := NewPath("/static/path")

	assert.Zero(t, len(p.ParamKeys))
	assert.Equal(t, "/static/path", p.Url())
}

type testCase struct {
	pattern     string
	params      []string
	expectedUrl string
}

func TestCases(t *testing.T) {
	tcs := []testCase{
		testCase{"/", nil, "/"},
		testCase{"//", nil, "//"},
		testCase{"/user/:user_id", []string{"joe"}, "/user/joe"},
		testCase{"/user/:user_id/", []string{"joe"}, "/user/joe/"},
		testCase{"/blog/:blog_id/post/:post_id", []string{"123", "456"}, "/blog/123/post/456"},
		testCase{"/:a/:b/:c", []string{"a", "b", "c"}, "/a/b/c"},
		testCase{"/:a/:b/:c//", []string{"a", "b", "c"}, "/a/b/c//"},
		testCase{"////:a///:b/:c//", []string{"a", "b", "c"}, "////a///b/c//"},
	}

	for _, tc := range tcs {
		tesTestCase(t, tc)
	}
}

func tesTestCase(t *testing.T, tc testCase) {
	p := NewPath(tc.pattern)
	assert.Equal(t, tc.expectedUrl, p.Url(tc.params...))
}

func TestNegativeCases(t *testing.T) {
	tcs := []testCase{
		testCase{"/", []string{"1"}, ""},
		testCase{"/:", []string{}, ""},
		testCase{"/:/", []string{"1"}, ""},
		testCase{"/:", []string{}, ""},
		testCase{"/:/", []string{"1"}, ""},
		testCase{"/:param", []string{}, ""},
		testCase{"/:param", []string{"1", "2"}, ""},
		testCase{"/:param1/:param2", []string{"1"}, ""},
		testCase{"/::invalid", []string{}, ""},
		testCase{"/:inva*lid", []string{}, ""},
		testCase{"/invalid_wildcard_position/*filepath/:dummy_var", []string{}, ""},
	}

	for _, tc := range tcs {
		testNegativeTestCase(t, tc)
	}
}

func testNegativeTestCase(t *testing.T, tc testCase) {
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()

		p := NewPath(tc.pattern)
		p.Url(tc.params...)

	}()
	assert.True(t, panicked, "expected negative test case to cause panic")
}
