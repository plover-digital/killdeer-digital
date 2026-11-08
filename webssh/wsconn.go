//go:build js && wasm

package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"syscall/js"
	"time"
)

// wsConn adapts a browser WebSocket to net.Conn so x/crypto/ssh can run over
// it. Every SSH byte travels as a binary WebSocket message; the relay turns
// those back into a TCP stream to the gateway.
type wsConn struct {
	ws     js.Value
	funcs  []js.Func
	events []string

	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	err    error // set once the socket closes or fails
	closed bool
	dead   chan struct{} // closed when err is first set
}

// dialWebSocket opens url and blocks until the socket is open or fails.
func dialWebSocket(url string) (*wsConn, error) {
	c := &wsConn{dead: make(chan struct{})}
	c.cond = sync.NewCond(&c.mu)

	ws := js.Global().Get("WebSocket").New(url)
	ws.Set("binaryType", "arraybuffer")
	c.ws = ws

	opened := make(chan error, 1)
	var once sync.Once
	signal := func(err error) { once.Do(func() { opened <- err }) }

	c.on("open", func(js.Value) { signal(nil) })
	c.on("message", func(ev js.Value) {
		data := js.Global().Get("Uint8Array").New(ev.Get("data"))
		b := make([]byte, data.Get("length").Int())
		js.CopyBytesToGo(b, data)
		c.mu.Lock()
		c.buf = append(c.buf, b...)
		c.cond.Broadcast()
		c.mu.Unlock()
	})
	c.on("error", func(js.Value) {
		signal(errors.New("could not reach the killdeer relay"))
		c.fail(errors.New("relay connection error"))
	})
	c.on("close", func(ev js.Value) {
		signal(errors.New("the killdeer relay closed the connection"))
		c.fail(io.EOF)
		// The socket fires nothing after close, so its callbacks can go.
		c.release()
	})

	if err := <-opened; err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *wsConn) on(event string, fn func(js.Value)) {
	f := js.FuncOf(func(this js.Value, args []js.Value) any {
		fn(args[0])
		return nil
	})
	c.funcs = append(c.funcs, f)
	c.events = append(c.events, event)
	c.ws.Call("addEventListener", event, f)
}

func (c *wsConn) fail(err error) {
	c.mu.Lock()
	if c.err == nil {
		c.err = err
		close(c.dead)
	}
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *wsConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.buf) == 0 && c.err == nil {
		c.cond.Wait()
	}
	if len(c.buf) == 0 {
		return 0, c.err
	}
	n := copy(p, c.buf)
	c.buf = c.buf[n:]
	return n, nil
}

func (c *wsConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	if err != nil {
		return 0, err
	}
	arr := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(arr, p)
	c.ws.Call("send", arr)
	return len(p), nil
}

func (c *wsConn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	if c.err == nil {
		c.err = net.ErrClosed
		close(c.dead)
	}
	c.cond.Broadcast()
	c.mu.Unlock()

	// The close event releases the callbacks once the browser is done.
	c.ws.Call("close")
	return nil
}

func (c *wsConn) release() {
	for i, f := range c.funcs {
		c.ws.Call("removeEventListener", c.events[i], f)
		f.Release()
	}
	c.funcs, c.events = nil, nil
}

type wsAddr string

func (a wsAddr) Network() string { return "websocket" }
func (a wsAddr) String() string  { return string(a) }

func (c *wsConn) LocalAddr() net.Addr                { return wsAddr("browser") }
func (c *wsConn) RemoteAddr() net.Addr               { return wsAddr(c.ws.Get("url").String()) }
func (c *wsConn) SetDeadline(t time.Time) error      { return nil }
func (c *wsConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *wsConn) SetWriteDeadline(t time.Time) error { return nil }
