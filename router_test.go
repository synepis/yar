package yar

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	router := New()

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMethodNotAllowed(t *testing.T) {
	// Arrange
	router := New()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {})
	r, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, r)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestSimplePath(t *testing.T) {
	// Arrange
	router := New()
	router.Get("/simplepath", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Served simple path"))
	})
	r, _ := http.NewRequest("GET", "/simplepath", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, r)

	// Assert
	output, _ := ioutil.ReadAll(w.Result().Body)
	assert.Equal(t, "Served simple path", string(output))
}

func TestSimplePathWithDifferentMethods(t *testing.T) {
	// Arrange
	router := New()
	router.Get("/simplepath", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET simple path"))
	})
	router.Post("/simplepath", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST simple path"))
	})
	rGet, _ := http.NewRequest("GET", "/simplepath", nil)
	wGet := httptest.NewRecorder()
	rPost, _ := http.NewRequest("POST", "/simplepath", nil)
	wPost := httptest.NewRecorder()

	// Act
	router.ServeHTTP(wGet, rGet)
	router.ServeHTTP(wPost, rPost)

	// Assert
	getOutput, _ := ioutil.ReadAll(wGet.Result().Body)
	postOutput, _ := ioutil.ReadAll(wPost.Result().Body)
	assert.Equal(t, "GET simple path", string(getOutput))
	assert.Equal(t, "POST simple path", string(postOutput))
}

func TestPathWithParamChars(t *testing.T) {
	// Arrange
	router := New()
	router.Get("/user/:id", func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r, "id")
		resp := fmt.Sprintf("Called with %s", id)
		w.Write([]byte(resp))
	})
	router.Get("/static/*filepath", func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r, "filepath")
		resp := fmt.Sprintf("Called with %s", id)
		w.Write([]byte(resp))
	})
	rGet1, _ := http.NewRequest("GET", "/user/:parameter", nil)
	wGet1 := httptest.NewRecorder()
	rGet2, _ := http.NewRequest("GET", "/static/*wildcard", nil)
	wGet2 := httptest.NewRecorder()

	// Act
	router.ServeHTTP(wGet1, rGet1)
	router.ServeHTTP(wGet2, rGet2)

	// Assert
	getOutput1, _ := ioutil.ReadAll(wGet1.Result().Body)
	getOutput2, _ := ioutil.ReadAll(wGet2.Result().Body)
	assert.Equal(t, "Called with :parameter", string(getOutput1))
	assert.Equal(t, "Called with *wildcard", string(getOutput2))
}

func TestParameterPathWithDifferentMethods(t *testing.T) {
	// Arrange
	router := New()
	router.Get("/user/:id", func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r, "id")
		resp := fmt.Sprintf("GET called with %s", id)
		w.Write([]byte(resp))
	})
	router.Post("/user/:id", func(w http.ResponseWriter, r *http.Request) {
		id := GetParam(r, "id")
		resp := fmt.Sprintf("POST called with %s", id)
		w.Write([]byte(resp))
	})
	rGet, _ := http.NewRequest("GET", "/user/1", nil)
	wGet := httptest.NewRecorder()
	rPost, _ := http.NewRequest("POST", "/user/2", nil)
	wPost := httptest.NewRecorder()

	// Act
	router.ServeHTTP(wGet, rGet)
	router.ServeHTTP(wPost, rPost)

	// Assert
	getOutput, _ := ioutil.ReadAll(wGet.Result().Body)
	postOutput, _ := ioutil.ReadAll(wPost.Result().Body)
	assert.Equal(t, "GET called with 1", string(getOutput))
	assert.Equal(t, "POST called with 2", string(postOutput))
}

func TestCustomNotFoundAddHandler(t *testing.T) {
	// Arrange
	router := New()
	router.notFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Custom Not Found"))
	})
	r, _ := http.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, r)

	// Assert
	output, _ := ioutil.ReadAll(w.Result().Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "Custom Not Found", string(output))
}

func TestCustomMethodNotAllowedAddHandler(t *testing.T) {
	// Arrange
	router := New()
	router.methodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Custom Method Not Allowed"))
	})
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {})
	r, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, r)

	// Assert
	output, _ := ioutil.ReadAll(w.Result().Body)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "Custom Method Not Allowed", string(output))
}

func TestParametersGetSet(t *testing.T) {
	// Arrange
	blogId, postId := "", ""

	router := New()
	router.Get("/blog/:blog_id/post/:post_id", func(w http.ResponseWriter, r *http.Request) {
		// errs := make([]error, 3)
		blogId = GetParam(r, "blog_id")
		postId = GetParam(r, "post_id")
		// params = GetParams(r)
		// for _, err := range errs {
		// 	assert.Nil(t, err, "expected to be no errors while reading parameters")
		// }
	})
	r, _ := http.NewRequest("GET", "/blog/123/post/456", nil)
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, r)

	// Assert
	assert.Equal(t, "123", blogId)
	assert.Equal(t, "456", postId)
	// assert.Equal(t, "123", params["blog_id"])
	// assert.Equal(t, "456", params["post_id"])
}

func BenchmarkStaticPath(b *testing.B) {
	router := New()

	router.Get("/static/path", func(w http.ResponseWriter, r *http.Request) {
	})

	r, _ := http.NewRequest("GET", "/static/path", nil)
	benchRequest(b, router, r)
}

func BenchmarkSingleParam(b *testing.B) {
	router := New()

	router.Get("/user/:name", func(w http.ResponseWriter, r *http.Request) {
	})

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}

func BenchmarkSingleParamWrite(b *testing.B) {
	router := New()

	router.Get("/user/:name", func(w http.ResponseWriter, r *http.Request) {
		param := GetParam(r, "name")
		io.WriteString(w, param)
	})

	r, _ := http.NewRequest("GET", "/user/gordon", nil)
	benchRequest(b, router, r)
}

func BenchmarkFiveParams(b *testing.B) {
	router := New()

	router.Get("/part1/:param1/part2/:param2/part3/:param3/part4/:param4/part5/:param5/",
		func(w http.ResponseWriter, r *http.Request) {})

	r, _ := http.NewRequest("GET", "/part1/part1/part2/part2/part3/part3/part4/part4/part5/part5", nil)
	benchRequest(b, router, r)
}

func BenchmarkFiveParamsWrite(b *testing.B) {
	router := New()

	router.Get("/part1/:param1/part2/:param2/part3/:param3/part4/:param4/part5/:param5/",
		func(w http.ResponseWriter, r *http.Request) {
			param1 := GetParam(r, "param1")
			param2 := GetParam(r, "param2")
			param3 := GetParam(r, "param3")
			param4 := GetParam(r, "param4")
			param5 := GetParam(r, "param5")
			io.WriteString(w, param1+param2+param3+param4+param5)
		})

	r, _ := http.NewRequest("GET", "/part1/part1/part2/part2/part3/part3/part4/part4/part5/part5", nil)
	benchRequest(b, router, r)
}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery
	r.RequestURI = u.RequestURI()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}
