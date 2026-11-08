package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

// terminalCSP is stricter than the homepage's: no third-party scripts, and
// 'wasm-unsafe-eval' so the SSH core can be compiled. connect-src covers the
// wasm fetch and the WebSocket relay on this origin.
const terminalCSP = "default-src 'none'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; " +
	"font-src 'self'; img-src 'self' data:; connect-src 'self' wss://killdeer.digital; " +
	"base-uri 'none'; form-action 'none'; frame-ancestors 'none'"

type terminalFile struct {
	body        []byte
	etag        string
	contentType string
}

type terminalHandler struct {
	files map[string]terminalFile
}

// newTerminalHandler loads the embedded /ssh/ files once and tags each with a
// content hash, so browsers revalidate cheaply instead of re-downloading the
// multi-megabyte SSH core.
func newTerminalHandler() *terminalHandler {
	sub, err := fs.Sub(sshFiles, "static/ssh")
	if err != nil {
		log.Fatal(err)
	}
	h := &terminalHandler{files: map[string]terminalFile{}}
	err = fs.WalkDir(sub, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		body, err := fs.ReadFile(sub, name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		h.files[name] = terminalFile{
			body:        body,
			etag:        `"` + hex.EncodeToString(sum[:12]) + `"`,
			contentType: terminalContentType(name),
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := h.files["killdeer-ssh.wasm"]; !ok {
		log.Printf("warning: static/ssh/killdeer-ssh.wasm is missing; run scripts/build-ssh-wasm.sh")
	}
	return h
}

func terminalContentType(name string) string {
	switch path.Ext(name) {
	case ".wasm":
		return "application/wasm"
	case ".mjs", ".js":
		return "text/javascript; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	}
	if t := mime.TypeByExtension(path.Ext(name)); t != "" {
		return t
	}
	return "application/octet-stream"
}

func (h *terminalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{Message: "Use GET or HEAD."})
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/ssh/")
	if name == "" {
		name = "index.html"
	}
	f, ok := h.files[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Security-Policy", terminalCSP)
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Content-Type", f.contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", f.etag)
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(f.body))
}
