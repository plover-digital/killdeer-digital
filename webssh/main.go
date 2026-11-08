//go:build js && wasm

// Command webssh is the SSH core of the killdeer.digital browser terminal.
// It is compiled to WebAssembly and runs entirely in the visitor's browser:
// the SSH handshake, the host key check and the password exchange all happen
// here, so the WebSocket relay only ever carries encrypted SSH bytes.
//
// JavaScript calls globalThis.killdeerSSH.connect(options) and gets back a
// session handle. See static/ssh/terminal.js for the caller.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"syscall/js"

	"golang.org/x/crypto/ssh"
)

// hostKey is the customer SSH gateway's public host key. Connections to any
// other key are refused, so a hostile relay cannot impersonate killdeer.
const hostKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEc3eNXm1JYCbdO+nkeyFtsKDVhmVLWPHz3rvASM/W87"

const sshHostname = "killdeer.digital"

func main() {
	pinned, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
	if err != nil {
		panic(err)
	}
	api := js.Global().Get("Object").New()
	api.Set("hostKeyFingerprint", ssh.FingerprintSHA256(pinned))
	api.Set("connect", js.FuncOf(func(this js.Value, args []js.Value) any {
		return connect(pinned, args[0])
	}))
	js.Global().Set("killdeerSSH", api)
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("killdeerssh:ready"))
	select {}
}

// session is one browser SSH connection.
type session struct {
	opts js.Value
	ctx  context.Context
	conn *wsConn // set once the relay is open

	input  *queue
	cancel context.CancelFunc

	prompted bool // the gateway asked at least one sign-in question

	mu      sync.Mutex
	sshSess *ssh.Session
	cols    int
	rows    int
}

// connect starts a connection in the background and returns its JS handle
// right away: {write(data), resize(cols, rows), close()}.
//
// options: url, username, cols, rows, and callbacks onStatus(stage, detail),
// onPrompt(question, echo) -> Promise<string>, onOutput(Uint8Array) and
// onClose(errorMessage).
func connect(pinned ssh.PublicKey, opts js.Value) js.Value {
	ctx, cancel := context.WithCancel(context.Background())
	s := &session{
		opts:   opts,
		ctx:    ctx,
		input:  newQueue(),
		cancel: cancel,
		cols:   opts.Get("cols").Int(),
		rows:   opts.Get("rows").Int(),
	}

	var funcs []js.Func
	fn := func(f func(args []js.Value)) js.Func {
		jf := js.FuncOf(func(this js.Value, args []js.Value) any {
			f(args)
			return nil
		})
		funcs = append(funcs, jf)
		return jf
	}
	handle := js.Global().Get("Object").New()
	handle.Set("write", fn(func(args []js.Value) {
		if args[0].Type() == js.TypeString {
			s.input.push([]byte(args[0].String()))
			return
		}
		b := make([]byte, args[0].Get("length").Int())
		js.CopyBytesToGo(b, args[0])
		s.input.push(b)
	}))
	handle.Set("resize", fn(func(args []js.Value) {
		go s.resize(args[0].Int(), args[1].Int())
	}))
	handle.Set("close", fn(func(args []js.Value) { cancel() }))

	go func() {
		err := s.run(ctx, pinned)
		cancel()
		msg := ""
		if err != nil && !errors.Is(err, context.Canceled) {
			msg = err.Error()
		}
		s.opts.Call("onClose", msg)
		for _, f := range funcs {
			f.Release()
		}
	}()
	return handle
}

func (s *session) status(stage, detail string) {
	s.opts.Call("onStatus", stage, detail)
}

func (s *session) run(ctx context.Context, pinned ssh.PublicKey) error {
	s.status("relay", "")
	conn, err := dialWebSocket(s.opts.Get("url").String())
	if err != nil {
		return err
	}
	context.AfterFunc(ctx, func() { conn.Close() })
	s.conn = conn
	s.status("relay-ok", "")

	config := &ssh.ClientConfig{
		User: s.opts.Get("username").String(),
		Auth: []ssh.AuthMethod{
			ssh.RetryableAuthMethod(ssh.KeyboardInteractive(s.challenge), 5),
		},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			if !bytes.Equal(key.Marshal(), pinned.Marshal()) {
				return fmt.Errorf("host key mismatch: got %s, want %s", ssh.FingerprintSHA256(key), ssh.FingerprintSHA256(pinned))
			}
			s.status("hostkey-ok", ssh.FingerprintSHA256(key))
			return nil
		},
		HostKeyAlgorithms: []string{pinned.Type()},
		BannerCallback: func(message string) error {
			s.output([]byte(message))
			return nil
		},
	}
	s.status("handshake", "")
	c, chans, reqs, err := ssh.NewClientConn(conn, sshHostname, config)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return s.friendlyError(err)
	}
	client := ssh.NewClient(c, chans, reqs)
	defer client.Close()
	s.status("auth-ok", "")

	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	s.mu.Lock()
	cols, rows := s.cols, s.rows
	s.mu.Unlock()
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 38400, ssh.TTY_OP_OSPEED: 38400}
	if err := sess.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		return err
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	sess.Stdout = writerFunc(s.output)
	sess.Stderr = writerFunc(s.output)
	if err := sess.Shell(); err != nil {
		return err
	}
	s.mu.Lock()
	s.sshSess = sess
	s.mu.Unlock()
	s.status("shell", "")

	go func() {
		for {
			b, ok := s.input.pop(ctx)
			if !ok {
				return
			}
			if _, err := stdin.Write(b); err != nil {
				return
			}
		}
	}()

	done := make(chan error, 1)
	go func() { done <- sess.Wait() }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		var exitErr *ssh.ExitMissingError
		if err == nil || errors.As(err, &exitErr) || errors.Is(err, io.EOF) {
			return nil
		}
		var ee *ssh.ExitError
		if errors.As(err, &ee) {
			return nil
		}
		return err
	}
}

// challenge relays each keyboard-interactive question (the password, or the
// signup and verification dialogue) to the page and waits for the answer.
func (s *session) challenge(name, instruction string, questions []string, echos []bool) ([]string, error) {
	if len(questions) > 0 {
		s.prompted = true
	}
	if name != "" {
		s.output([]byte("  " + name + "\r\n"))
	}
	if instruction != "" {
		s.output([]byte("  " + instruction + "\r\n"))
	}
	answers := make([]string, len(questions))
	for i, q := range questions {
		a, err := s.prompt(q, echos[i])
		if err != nil {
			return nil, err
		}
		answers[i] = a
	}
	return answers, nil
}

func (s *session) prompt(question string, echo bool) (string, error) {
	type result struct {
		answer string
		err    error
	}
	ch := make(chan result, 1)
	// The promise may settle after we stop waiting (the connection dropped),
	// so the callbacks release themselves rather than being released here:
	// calling a released js.Func would crash the whole runtime.
	var then, catch js.Func
	settle := func(r result) {
		ch <- r
		then.Release()
		catch.Release()
	}
	then = js.FuncOf(func(this js.Value, args []js.Value) any {
		settle(result{answer: args[0].String()})
		return nil
	})
	catch = js.FuncOf(func(this js.Value, args []js.Value) any {
		settle(result{err: errors.New("sign-in cancelled")})
		return nil
	})
	s.opts.Call("onPrompt", question, echo).Call("then", then, catch)
	select {
	case r := <-ch:
		return r.answer, r.err
	case <-s.conn.dead:
		return "", errors.New("connection lost")
	case <-s.ctx.Done():
		return "", s.ctx.Err()
	}
}

func (s *session) resize(cols, rows int) {
	s.mu.Lock()
	s.cols, s.rows = cols, rows
	sess := s.sshSess
	s.mu.Unlock()
	if sess != nil {
		_ = sess.WindowChange(rows, cols)
	}
}

func (s *session) output(b []byte) {
	arr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(arr, b)
	s.opts.Call("onOutput", arr)
}

func (s *session) friendlyError(err error) error {
	msg := err.Error()
	// The gateway refuses keyboard-interactive without asking anything when
	// the account has no password (it signs in with an SSH key only).
	// x/crypto reports that as "unexpected message type 51 (expected 60)".
	refused := strings.Contains(msg, "unexpected message type 51") || strings.Contains(msg, "unable to authenticate")
	switch {
	case refused && !s.prompted:
		user := s.opts.Get("username").String()
		// A core-side error refuses the same way, so this is a likely
		// explanation, not a certain one.
		return fmt.Errorf("killdeer didn't ask %s for a password. That usually means the account signs in with an SSH key. From a computer that has the key, run:\n  ssh %s@killdeer.digital\nIf this account does have a password, try again in a moment.", user, user)
	case refused:
		return errors.New("sign-in failed")
	case strings.Contains(msg, "host key mismatch"):
		return fmt.Errorf("refusing to connect: %v", err)
	}
	return err
}

type writerFunc func([]byte)

func (w writerFunc) Write(p []byte) (int, error) {
	b := make([]byte, len(p))
	copy(b, p)
	w(b)
	return len(p), nil
}
