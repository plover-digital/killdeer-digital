package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTerminalHandler(t *testing.T) {
	h := newTerminalHandler()

	get := func(path string, header ...string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		for i := 0; i+1 < len(header); i += 2 {
			req.Header.Set(header[i], header[i+1])
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	page := get("/ssh/")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "terminal.js") {
		t.Fatalf("GET /ssh/ = %d, want the terminal page", page.Code)
	}
	csp := page.Header().Get("Content-Security-Policy")
	for _, want := range []string{"'wasm-unsafe-eval'", "frame-ancestors 'none'", "default-src 'none'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP %q is missing %s", csp, want)
		}
	}
	if strings.Contains(csp, "unsafe-inline' https") || strings.Contains(csp, "analytics") {
		t.Errorf("terminal CSP allows third-party scripts: %q", csp)
	}

	js := get("/ssh/terminal.js")
	if ct := js.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/javascript") {
		t.Errorf("terminal.js Content-Type = %q", ct)
	}
	etag := js.Header().Get("ETag")
	if etag == "" {
		t.Fatal("terminal.js has no ETag")
	}
	if rec := get("/ssh/terminal.js", "If-None-Match", etag); rec.Code != http.StatusNotModified {
		t.Errorf("revalidation = %d, want 304", rec.Code)
	}

	for _, path := range []string{"/ssh/missing", "/ssh/../main.go", "/ssh/_ws/killdeer.digital"} {
		if rec := get(path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/ssh/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /ssh/ = %d, want 405", rec.Code)
	}
}
