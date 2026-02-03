package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStack(t *testing.T) {
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-M1", "true")
			next.ServeHTTP(w, r)
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-M2", "true")
			next.ServeHTTP(w, r)
		})
	}

	stack := NewStack(m1, m2)

	assert.NotNil(t, stack)
	assert.Equal(t, 2, stack.Count())
}

func TestStackThen(t *testing.T) {
	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1-before")
			next.ServeHTTP(w, r)
			order = append(order, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2-before")
			next.ServeHTTP(w, r)
			order = append(order, "m2-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	stack := NewStack(m1, m2)
	wrappedHandler := stack.Then(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Expected execution order
	expected := []string{
		"m1-before",
		"m2-before",
		"handler",
		"m2-after",
		"m1-after",
	}

	assert.Equal(t, expected, order)
}

func TestStackUse(t *testing.T) {
	stack := NewStack()
	assert.Equal(t, 0, stack.Count())

	m1 := func(next http.Handler) http.Handler { return next }
	stack.Use(m1)
	assert.Equal(t, 1, stack.Count())

	m2 := func(next http.Handler) http.Handler { return next }
	stack.Use(m2)
	assert.Equal(t, 2, stack.Count())
}

func TestStackClone(t *testing.T) {
	m1 := func(next http.Handler) http.Handler { return next }
	m2 := func(next http.Handler) http.Handler { return next }

	original := NewStack(m1, m2)
	clone := original.Clone()

	assert.Equal(t, original.Count(), clone.Count())

	// Modifying the clone should not affect the original
	m3 := func(next http.Handler) http.Handler { return next }
	clone.Use(m3)

	assert.Equal(t, 2, original.Count())
	assert.Equal(t, 3, clone.Count())
}

func TestChain(t *testing.T) {
	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1")
			next.ServeHTTP(w, r)
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	chained := Chain(m1, m2)
	wrappedHandler := chained(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	expected := []string{"m1", "m2", "handler"}
	assert.Equal(t, expected, order)
}

func TestConditional(t *testing.T) {
	var executed bool

	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executed = true
			next.ServeHTTP(w, r)
		})
	}

	// Condition true
	condition := func(r *http.Request) bool {
		return r.URL.Path == "/test"
	}

	conditional := Conditional(condition, m)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Test avec condition true
	executed = false
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	conditional(handler).ServeHTTP(rec, req)
	assert.True(t, executed)

	// Test avec condition false
	executed = false
	req = httptest.NewRequest("GET", "/other", nil)
	rec = httptest.NewRecorder()
	conditional(handler).ServeHTTP(rec, req)
	assert.False(t, executed)
}

func TestSkipPaths(t *testing.T) {
	var executed bool

	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executed = true
			next.ServeHTTP(w, r)
		})
	}

	skipPaths := []string{"/health", "/metrics"}
	wrapped := SkipPaths(skipPaths, m)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Path skipped
	executed = false
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.False(t, executed)

	// Path not skipped
	executed = false
	req = httptest.NewRequest("GET", "/api", nil)
	rec = httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.True(t, executed)
}

func TestOnlyPaths(t *testing.T) {
	var executed bool

	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executed = true
			next.ServeHTTP(w, r)
		})
	}

	onlyPaths := []string{"/admin", "/api"}
	wrapped := OnlyPaths(onlyPaths, m)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Path in list
	executed = false
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.True(t, executed)

	// Path not in list
	executed = false
	req = httptest.NewRequest("GET", "/public", nil)
	rec = httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.False(t, executed)
}

func TestOnlyMethods(t *testing.T) {
	var executed bool

	m := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executed = true
			next.ServeHTTP(w, r)
		})
	}

	onlyMethods := []string{"POST", "PUT"}
	wrapped := OnlyMethods(onlyMethods, m)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Method in list
	executed = false
	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.True(t, executed)

	// Method not in list
	executed = false
	req = httptest.NewRequest("GET", "/", nil)
	rec = httptest.NewRecorder()
	wrapped(handler).ServeHTTP(rec, req)
	assert.False(t, executed)
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.Status())

	// Test Write
	n, err := rw.Write([]byte("Hello World"))
	require.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, 11, rw.Size())

	// Test multiple writes
	rw.Write([]byte(" Test"))
	assert.Equal(t, 16, rw.Size())
}

func TestResponseWriterDefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// Write without WriteHeader should use 200
	rw.Write([]byte("test"))
	assert.Equal(t, http.StatusOK, rw.Status())
}

func TestResponseWriterMultipleWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// First WriteHeader
	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.Status())

	// Second WriteHeader should not change status
	rw.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusCreated, rw.Status())
}

func TestApply(t *testing.T) {
	var order []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m1")
			next.ServeHTTP(w, r)
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "m2")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	wrapped := Apply(handler, m1, m2)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	expected := []string{"m1", "m2", "handler"}
	assert.Equal(t, expected, order)
}

func BenchmarkStack(b *testing.B) {
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	stack := NewStack(m1, m1, m1)
	wrappedHandler := stack.Then(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
	}
}

func BenchmarkChain(b *testing.B) {
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chained := Chain(m1, m1, m1)
	wrappedHandler := chained(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
	}
}

func ExampleStack() {
	// Create a middleware stack
	stack := NewStack(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Middleware 1")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Middleware 2")
				next.ServeHTTP(w, r)
			})
		},
	)

	// Apply to a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handler")
	})

	wrappedHandler := stack.Then(handler)

	// Test
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	// Output:
	// Middleware 1
	// Middleware 2
	// Handler
}
