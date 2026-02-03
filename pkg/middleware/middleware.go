package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/samber/lo"
)

// Middleware is the standard Chi-style signature.
type Middleware func(http.Handler) http.Handler

// Stack allows composing multiple middlewares.
type Stack struct {
	middlewares []Middleware
}

// NewStack creates a new middleware stack.
func NewStack(middlewares ...Middleware) *Stack {
	return &Stack{
		middlewares: middlewares,
	}
}

// Use adds a middleware to the stack.
func (s *Stack) Use(m Middleware) *Stack {
	s.middlewares = append(s.middlewares, m)
	return s
}

// UseMultiple adds multiple middlewares to the stack.
func (s *Stack) UseMultiple(middlewares ...Middleware) *Stack {
	s.middlewares = append(s.middlewares, middlewares...)
	return s
}

// Then applies all middlewares to a handler.
func (s *Stack) Then(h http.Handler) http.Handler {
	return lo.ReduceRight(s.middlewares, func(handler http.Handler, m Middleware, _ int) http.Handler {
		return m(handler)
	}, h)
}

// ThenFunc applies all middlewares to a HandlerFunc.
func (s *Stack) ThenFunc(fn http.HandlerFunc) http.Handler {
	return s.Then(fn)
}

// Clone creates a copy of the stack.
func (s *Stack) Clone() *Stack {
	return &Stack{
		middlewares: append([]Middleware{}, s.middlewares...),
	}
}

// Count returns the number of middlewares in the stack.
func (s *Stack) Count() int {
	return len(s.middlewares)
}

// Clear empties the stack.
func (s *Stack) Clear() *Stack {
	s.middlewares = []Middleware{}
	return s
}

// Chain composes multiple middlewares into one.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		return lo.ReduceRight(middlewares, func(handler http.Handler, m Middleware, _ int) http.Handler {
			return m(handler)
		}, next)
	}
}

// Compose is an alias for Chain (compatibility).
func Compose(middlewares ...Middleware) Middleware {
	return Chain(middlewares...)
}

// Apply applies a middleware to a handler (helper).
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return NewStack(middlewares...).Then(h)
}

// ApplyFunc applies middlewares to a HandlerFunc (helper).
func ApplyFunc(fn http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Apply(fn, middlewares...)
}

// Conditional returns a conditional middleware.
func Conditional(condition func(*http.Request) bool, m Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if condition(r) {
				m(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// SkipPaths returns a middleware that skips certain paths.
func SkipPaths(paths []string, m Middleware) Middleware {
	return Conditional(func(r *http.Request) bool {
		return !lo.Contains(paths, r.URL.Path)
	}, m)
}

// OnlyPaths returns a middleware that applies only to certain paths.
func OnlyPaths(paths []string, m Middleware) Middleware {
	return Conditional(func(r *http.Request) bool {
		return lo.Contains(paths, r.URL.Path)
	}, m)
}

// OnlyMethods returns a middleware that applies only to certain HTTP methods.
func OnlyMethods(methods []string, m Middleware) Middleware {
	return Conditional(func(r *http.Request) bool {
		return lo.Contains(methods, r.Method)
	}, m)
}

// Group allows grouping routes with common middlewares.
type Group struct {
	stack *Stack
}

// NewGroup creates a new route group.
func NewGroup(middlewares ...Middleware) *Group {
	return &Group{
		stack: NewStack(middlewares...),
	}
}

// Use adds a middleware to the group.
func (g *Group) Use(m Middleware) *Group {
	g.stack.Use(m)
	return g
}

// Handle registers a handler with the group's middlewares.
func (g *Group) Handle(pattern string, h http.Handler, mux *http.ServeMux) {
	mux.Handle(pattern, g.stack.Then(h))
}

// HandleFunc registers a HandlerFunc with the group's middlewares.
func (g *Group) HandleFunc(pattern string, fn http.HandlerFunc, mux *http.ServeMux) {
	mux.Handle(pattern, g.stack.ThenFunc(fn))
}

// responseWriter wrapper to capture status and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	wrote  bool
}

// NewResponseWriter creates a ResponseWriter wrapper.
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		size:           0,
		wrote:          false,
	}
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(status int) {
	if !rw.wrote {
		rw.status = status
		rw.wrote = true
		rw.ResponseWriter.WriteHeader(status)
	}
}

// Write captures the response size.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wrote {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Status returns the response status code.
func (rw *responseWriter) Status() int {
	return rw.status
}

// Size returns the response size in bytes.
func (rw *responseWriter) Size() int {
	return rw.size
}

// Unwrap allows access to the original ResponseWriter.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Flush implements http.Flusher.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// Push implements http.Pusher (HTTP/2 Server Push).
func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := rw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return fmt.Errorf("ResponseWriter does not implement http.Pusher")
}
