package engine

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	_ "net/http/pprof" // register pprof handlers on DefaultServeMux
	"strings"
)

// ---------------------------------------------------------------------------
// pprof — debug profiling endpoint
// ---------------------------------------------------------------------------

// EnablePprof mounts the Go pprof profiling endpoints at /debug/pprof/
// on the given mux. Call this only in development or behind auth.
//
// Usage:
//
//	panel.WithMiddleware(engine.PprofAuthMiddleware("secret-token"))
//	engine.EnablePprof(mux)
func EnablePprof(mux *http.ServeMux) {
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
}

// PprofAuthMiddleware restricts /debug/pprof/ to requests with the correct
// Bearer token. Use this to protect pprof in staging environments.
func PprofAuthMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
				auth := r.Header.Get("Authorization")
				if auth != "Bearer "+token {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// ETag middleware — conditional GET support for HTML responses
// ---------------------------------------------------------------------------

// etagResponseWriter buffers the response body to compute an ETag.
type etagResponseWriter struct {
	http.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (e *etagResponseWriter) WriteHeader(code int) {
	e.status = code
}

func (e *etagResponseWriter) Write(b []byte) (int, error) {
	return e.buf.Write(b)
}

// ETagMiddleware computes a SHA-256 ETag for GET responses and returns
// 304 Not Modified when the client's If-None-Match matches.
// Only applied to 200 OK HTML responses to avoid buffering large exports.
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only GET/HEAD benefit from ETags
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		erw := &etagResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(erw, r)

		// Only cache HTML 200 responses
		ct := w.Header().Get("Content-Type")
		if erw.status != http.StatusOK || !strings.Contains(ct, "text/html") {
			if erw.status != http.StatusOK {
				w.WriteHeader(erw.status)
			}
			_, _ = w.Write(erw.buf.Bytes())
			return
		}

		body := erw.buf.Bytes()
		etag := fmt.Sprintf(`"%x"`, sha256.Sum256(body))

		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")

		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.WriteHeader(erw.status)
		_, _ = w.Write(body)
	})
}

// ---------------------------------------------------------------------------
// SecurityHeaders middleware — OWASP recommended headers
// ---------------------------------------------------------------------------

// SecurityHeadersMiddleware adds OWASP-recommended security headers to every response.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
