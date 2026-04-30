package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed static/index.html static/index.md static/llms.txt static/llms-full.txt static/ssh-help.txt static/sizes.txt static/os.txt static/robots.txt static/sitemap.xml static/social-card.html static/social-card.png static/api/v1/cli.json static/api/v1/sizes.json static/api/v1/images.json static/openapi.json static/.well-known/api-catalog static/.well-known/agent-skills/index.json static/.well-known/agent-skills/killdeer-cli/SKILL.md static/.well-known/agent-skills/killdeer-sizing/SKILL.md
var staticFiles embed.FS

const (
	defaultPort   = "8080"
	contentSignal = "search=yes, ai-input=yes, ai-train=no"
)

type config struct {
	Port string
}

type jsonResponse struct {
	Message string `json:"message"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/index.html", handleIndex)
	mux.HandleFunc("/index.md", handleIndexMarkdown)
	mux.HandleFunc("/llms.txt", handleLLMS)
	mux.HandleFunc("/.well-known/llms.txt", handleLLMS)
	mux.HandleFunc("/llms-full.txt", handleLLMSFull)
	mux.HandleFunc("/ssh-help.txt", handleSSHHelp)
	mux.HandleFunc("/sizes.txt", handleSizes)
	mux.HandleFunc("/os.txt", handleOS)
	mux.HandleFunc("/api/v1/cli.json", handleAPICLI)
	mux.HandleFunc("/api/v1/sizes.json", handleAPISizes)
	mux.HandleFunc("/api/v1/images.json", handleAPIImages)
	mux.HandleFunc("/openapi.json", handleOpenAPI)
	mux.HandleFunc("/.well-known/api-catalog", handleAPICatalog)
	mux.HandleFunc("/.well-known/agent-skills/index.json", handleAgentSkillsIndex)
	mux.HandleFunc("/.well-known/agent-skills/killdeer-cli/SKILL.md", handleKilldeerCLISkill)
	mux.HandleFunc("/.well-known/agent-skills/killdeer-sizing/SKILL.md", handleKilldeerSizingSkill)
	mux.HandleFunc("/robots.txt", handleRobots)
	mux.HandleFunc("/sitemap.xml", handleSitemap)
	mux.HandleFunc("/social-card.html", handleSocialCardHTML)
	mux.HandleFunc("/social-card.png", handleSocialCard)
	mux.HandleFunc("/healthz", handleHealth)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	addr := ":" + cfg.Port
	log.Printf("killdeer.digital site listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, withCommonHeaders(withRequestLog(mux))))
}

func loadConfig() (config, error) {
	cfg := config{
		Port: defaultPort,
	}

	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		cfg.Port = port
	}

	return cfg, nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}

	setDiscoveryHeaders(w)
	if acceptsMarkdown(r.Header.Get("Accept")) {
		serveEmbeddedFile(w, r, "static/index.md", "text/markdown; charset=utf-8")
		return
	}

	serveEmbeddedFile(w, r, "static/index.html", "text/html; charset=utf-8")
}

func handleIndexMarkdown(w http.ResponseWriter, r *http.Request) {
	setDiscoveryHeaders(w)
	serveEmbeddedFile(w, r, "static/index.md", "text/markdown; charset=utf-8")
}

func handleLLMS(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/llms.txt", "text/plain; charset=utf-8")
}

func handleLLMSFull(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/llms-full.txt", "text/plain; charset=utf-8")
}

func handleSSHHelp(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/ssh-help.txt", "text/plain; charset=utf-8")
}

func handleSizes(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/sizes.txt", "text/plain; charset=utf-8")
}

func handleOS(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/os.txt", "text/plain; charset=utf-8")
}

func handleAPICLI(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/api/v1/cli.json", "application/json; charset=utf-8")
}

func handleAPISizes(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/api/v1/sizes.json", "application/json; charset=utf-8")
}

func handleAPIImages(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/api/v1/images.json", "application/json; charset=utf-8")
}

func handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/openapi.json", "application/vnd.oai.openapi+json; charset=utf-8")
}

func handleAPICatalog(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/.well-known/api-catalog", "application/linkset+json; charset=utf-8")
}

func handleAgentSkillsIndex(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/.well-known/agent-skills/index.json", "application/json; charset=utf-8")
}

func handleKilldeerCLISkill(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/.well-known/agent-skills/killdeer-cli/SKILL.md", "text/markdown; charset=utf-8")
}

func handleKilldeerSizingSkill(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/.well-known/agent-skills/killdeer-sizing/SKILL.md", "text/markdown; charset=utf-8")
}

func handleRobots(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/robots.txt", "text/plain; charset=utf-8")
}

func handleSitemap(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/sitemap.xml", "application/xml; charset=utf-8")
}

func handleSocialCard(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/social-card.png", "image/png")
}

func handleSocialCardHTML(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedFile(w, r, "static/social-card.html", "text/html; charset=utf-8")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{Message: "Use GET or HEAD."})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte("ok\n"))
	}
}

func serveEmbeddedFile(w http.ResponseWriter, r *http.Request, path, contentType string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, http.StatusMethodNotAllowed, jsonResponse{Message: "Use GET or HEAD."})
		return
	}

	body, err := staticFiles.ReadFile(path)
	if err != nil {
		http.Error(w, "embedded asset missing", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	// These assets are tiny and change during active iteration, so favor freshness over caching.
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

func setDiscoveryHeaders(w http.ResponseWriter) {
	w.Header().Add("Link", `</index.md>; rel="alternate"; type="text/markdown"`)
	w.Header().Add("Link", `</llms.txt>; rel="alternate"; type="text/plain"`)
	w.Header().Add("Link", `</llms-full.txt>; rel="alternate"; type="text/plain"`)
	w.Header().Add("Link", `</ssh-help.txt>; rel="alternate"; type="text/plain"`)
	w.Header().Add("Link", `</sizes.txt>; rel="alternate"; type="text/plain"`)
	w.Header().Add("Link", `</os.txt>; rel="alternate"; type="text/plain"`)
	w.Header().Add("Link", `</sitemap.xml>; rel="alternate"; type="application/xml"`)
	w.Header().Add("Link", `</.well-known/api-catalog>; rel="api-catalog"; type="application/linkset+json"`)
	w.Header().Add("Link", `</.well-known/agent-skills/index.json>; rel="agent-skills"; type="application/json"`)
	w.Header().Add("Vary", "Accept")
}

func acceptsMarkdown(accept string) bool {
	return strings.Contains(strings.ToLower(accept), "text/markdown")
}

func writeJSON(w http.ResponseWriter, status int, payload jsonResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Signal", contentSignal)
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' https://app.mailjet.com https://analytics.plover.digital; img-src 'self' data: https://app.mailjet.com https://0607p.mjt.lu; connect-src 'self' https://app.mailjet.com https://0607p.mjt.lu https://analytics.plover.digital; frame-src 'self' https://0607p.mjt.lu; base-uri 'none'; form-action 'self' https://0607p.mjt.lu; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
