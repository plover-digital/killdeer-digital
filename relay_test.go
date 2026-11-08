package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// echoUpstream stands in for the SSH gateway.
func echoUpstream(t *testing.T) string {
	return fakeUpstream(t, func(c net.Conn) { _, _ = io.Copy(c, c) })
}

// sinkUpstream swallows everything, so a client only ever reads closes.
func sinkUpstream(t *testing.T) string {
	return fakeUpstream(t, func(c net.Conn) { _, _ = io.Copy(io.Discard, c) })
}

func fakeUpstream(t *testing.T, serve func(net.Conn)) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() { defer c.Close(); serve(c) }()
		}
	}()
	return ln.Addr().String()
}

func relayServer(t *testing.T, target string) (*relay, string) {
	t.Helper()
	rl := newRelay()
	rl.target = target
	srv := httptest.NewServer(rl)
	t.Cleanup(srv.Close)
	return rl, "ws" + strings.TrimPrefix(srv.URL, "http") + relayPath
}

func TestRelayCarriesBytes(t *testing.T) {
	_, url := relayServer(t, echoUpstream(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	conn := websocket.NetConn(ctx, c, websocket.MessageBinary)
	want := []byte("SSH-2.0-test\r\n")
	if _, err := conn.Write(want); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(want))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("relay returned %q, want %q", got, want)
	}
}

func TestRelayRefusesOtherOrigins(t *testing.T) {
	_, url := relayServer(t, echoUpstream(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h := http.Header{"Origin": {"https://evil.example"}}
	_, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: h})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin upgrade: err=%v resp=%v, want 403", err, resp)
	}
}

func TestRelayOnlyServesItsPath(t *testing.T) {
	rl := newRelay()
	rec := httptest.NewRecorder()
	rl.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ssh/_ws/elsewhere.example", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("other relay path = %d, want 404", rec.Code)
	}
}

func TestRelayCapsTotalConnections(t *testing.T) {
	rl := newRelay()
	for i := 0; i < relayMaxTotal; i++ {
		if !rl.acquire() {
			t.Fatalf("connection %d refused below the cap", i+1)
		}
	}
	if rl.acquire() {
		t.Fatal("connection above the cap was allowed")
	}
	rl.release()
	if !rl.acquire() {
		t.Fatal("slot was not released")
	}
}

func TestRelayRejectsOversizedMessages(t *testing.T) {
	_, url := relayServer(t, sinkUpstream(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()
	if err := c.Write(ctx, websocket.MessageBinary, make([]byte, 2*relayMaxMessage)); err != nil {
		t.Fatal(err)
	}
	// The relay must close the connection instead of forwarding it.
	if _, _, err := c.Read(ctx); websocket.CloseStatus(err) != websocket.StatusMessageTooBig {
		t.Fatalf("oversized message: got %v, want close status %d", err, websocket.StatusMessageTooBig)
	}
}

func TestClientAddressUsesCaddysHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, relayPath, nil)
	r.RemoteAddr = "10.89.0.2:5555"
	if got := clientAddress(r); got != "10.89.0.2" {
		t.Errorf("without header = %q", got)
	}
	r.Header.Set("X-Forwarded-For", "spoofed, 192.0.2.7")
	if got := clientAddress(r); got != "192.0.2.7" {
		t.Errorf("with header = %q, want the last hop Caddy added", got)
	}
}
