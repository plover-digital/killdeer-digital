package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// The browser terminal's WebSocket-to-SSH relay. Browsers cannot open raw TCP
// connections, so the WASM SSH client (webssh/) tunnels its SSH stream through
// this one WebSocket endpoint to the customer SSH gateway. The bytes are
// already SSH-encrypted end to end; the relay never sees passwords or
// session contents, and it only ever dials the configured gateway.

const (
	relayPath         = "/ssh/_ws/killdeer.digital"
	defaultSSHTarget  = "killdeer.digital:22"
	relayDialTimeout  = 10 * time.Second
	relayMaxMessage   = 1 << 20 // SSH packets are far smaller
	relayMaxTotal     = 256
	relayMaxPerClient = 8
)

type relay struct {
	target string
	dial   func(ctx context.Context, network, addr string) (net.Conn, error)

	mu        sync.Mutex
	total     int
	perClient map[string]int
}

func newRelay() *relay {
	target := strings.TrimSpace(os.Getenv("SSH_TARGET"))
	if target == "" {
		target = defaultSSHTarget
	}
	d := &net.Dialer{Timeout: relayDialTimeout, KeepAlive: 30 * time.Second}
	return &relay{target: target, dial: d.DialContext, perClient: map[string]int{}}
}

// clientAddress is the visitor's address as Caddy reports it. Caddy is the
// only thing that can reach this server, and it sets X-Forwarded-For itself.
func clientAddress(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.LastIndexByte(fwd, ','); i >= 0 {
			fwd = fwd[i+1:]
		}
		return strings.TrimSpace(fwd)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (rl *relay) acquire(client string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.total >= relayMaxTotal || rl.perClient[client] >= relayMaxPerClient {
		return false
	}
	rl.total++
	rl.perClient[client]++
	return true
}

func (rl *relay) release(client string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.total--
	if rl.perClient[client]--; rl.perClient[client] <= 0 {
		delete(rl.perClient, client)
	}
}

func (rl *relay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != relayPath {
		http.NotFound(w, r)
		return
	}
	client := clientAddress(r)
	if !rl.acquire(client) {
		http.Error(w, "too many terminal connections", http.StatusTooManyRequests)
		return
	}
	defer rl.release(client)

	// Accept refuses cross-origin upgrades: with no OriginPatterns, the
	// Origin header must match the Host the page was served from.
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer ws.CloseNow()
	ws.SetReadLimit(relayMaxMessage)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	dialCtx, dialCancel := context.WithTimeout(ctx, relayDialTimeout)
	upstream, err := rl.dial(dialCtx, "tcp", rl.target)
	dialCancel()
	if err != nil {
		log.Printf("terminal relay: client %s: dial %s: %v", client, rl.target, err)
		ws.Close(websocket.StatusTryAgainLater, "killdeer is unreachable right now")
		return
	}
	defer upstream.Close()

	started := time.Now()
	log.Printf("terminal relay: client %s connected", client)
	stream := websocket.NetConn(ctx, ws, websocket.MessageBinary)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			upstream.Close()
			stream.Close()
		})
	}
	go func() {
		defer stop()
		_, _ = io.Copy(upstream, stream)
	}()
	_, _ = io.Copy(stream, upstream)
	stop()
	log.Printf("terminal relay: client %s disconnected after %s", client, time.Since(started).Round(time.Second))
}
