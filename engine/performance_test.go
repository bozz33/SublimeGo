package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// BenchmarkGzipMiddlewarePool measures the gzip pool vs naive allocation.
func BenchmarkGzipMiddlewarePool(b *testing.B) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("Hello SublimeGo! ", 200)))
	})
	handler := gzipMiddleware(inner)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// BenchmarkGzipMiddlewareNoGzip measures the no-gzip fast path.
func BenchmarkGzipMiddlewareNoGzip(b *testing.B) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("Hello SublimeGo! ", 200)))
	})
	handler := gzipMiddleware(inner)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// BenchmarkSubdomainResolver measures tenant resolution from host header.
func BenchmarkSubdomainResolver(b *testing.B) {
	resolver := NewSubdomainResolver("example.com")
	for i := 0; i < 100; i++ {
		resolver.Register(&Tenant{
			ID:        strings.Repeat("t", 4),
			Subdomain: "tenant" + strings.Repeat("x", i%10),
		})
	}
	resolver.Register(&Tenant{ID: "acme", Subdomain: "acme"})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "acme.example.com"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		resolver.Resolve(r)
	}
}

// BenchmarkTenantMiddleware measures the full middleware chain cost.
func BenchmarkTenantMiddleware(b *testing.B) {
	resolver := NewSubdomainResolver("example.com")
	resolver.Register(&Tenant{ID: "acme", Subdomain: "acme"})

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := TenantMiddleware(resolver, false)(inner)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "acme.example.com"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// BenchmarkCRUDHandlerList measures the List dispatch path (no DB).
func BenchmarkCRUDHandlerList(b *testing.B) {
	res := NewSimpleResource("items", "Item", "Items")
	res.WithList(func(_ context.Context) ([]any, error) {
		return []any{}, nil
	})
	handler := NewCRUDHandler(res)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/items", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}
