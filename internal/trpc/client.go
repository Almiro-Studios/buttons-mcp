// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

// Package trpc is a minimal client for the tRPC-over-WebSocket API that the
// Bitfocus Buttons web UI uses (wss://<host>/api/wstrpc). Inputs and outputs
// use the superjson envelope {"json": ..., "meta": ...}.
package trpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// Client is a single authenticated WebSocket session.
type Client struct {
	URL      string
	Username string
	Password string
	Insecure bool // accept the self-signed certificate Buttons ships with

	http    *http.Client
	mu      sync.Mutex
	conn    *websocket.Conn
	nextID  atomic.Int64
	pending sync.Map // id -> chan response
}

type response struct {
	ID     int64 `json:"id"`
	Result *struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	} `json:"result"`
	Error json.RawMessage `json:"error"`
}

// Error is a tRPC error returned by the server.
type Error struct {
	Path    string
	Message string
	Code    string
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s (%s)", e.Path, e.Message, e.Code) }

func (c *Client) httpClient() *http.Client {
	if c.http == nil {
		jar, _ := cookiejar.New(nil)
		tr := &http.Transport{}
		if c.Insecure {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
		c.http = &http.Client{Transport: tr, Jar: jar, Timeout: 15 * time.Second}
	}
	return c.http
}

// httpBase turns wss://host/api/wstrpc into https://host/api/trpc.
func (c *Client) httpBase() string {
	u := strings.Replace(c.URL, "wss://", "https://", 1)
	u = strings.Replace(u, "ws://", "http://", 1)
	return strings.Replace(u, "/api/wstrpc", "/api/trpc", 1)
}

// login signs in over HTTP; the server answers with a session cookie that
// the WebSocket handshake then carries.
func (c *Client) login(ctx context.Context) error {
	body, _ := json.Marshal(map[string]any{"json": map[string]any{"username": c.Username, "password": c.Password}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.httpBase()+"/auth.localLogin", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var e struct {
			Error struct {
				JSON struct {
					Message string `json:"message"`
				} `json:"json"`
			} `json:"error"`
		}
		b, _ := io.ReadAll(res.Body)
		_ = json.Unmarshal(b, &e)
		if e.Error.JSON.Message != "" {
			return errors.New(e.Error.JSON.Message)
		}
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return nil
}

func (c *Client) ensure(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return nil
	}
	hc := c.httpClient()
	if c.Username != "" {
		if err := c.login(ctx); err != nil {
			return fmt.Errorf("login: %w", err)
		}
	}
	opts := &websocket.DialOptions{HTTPClient: &http.Client{Transport: hc.Transport}}
	if u, err := url.Parse(c.httpBase()); err == nil {
		var parts []string
		for _, ck := range hc.Jar.Cookies(u) {
			parts = append(parts, ck.Name+"="+ck.Value)
		}
		if len(parts) > 0 {
			opts.HTTPHeader = http.Header{"Cookie": {strings.Join(parts, "; ")}}
		}
	}
	conn, _, err := websocket.Dial(ctx, c.URL, opts)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.URL, err)
	}
	conn.SetReadLimit(64 << 20)
	c.conn = conn
	go c.readLoop(conn)
	return nil
}

func (c *Client) readLoop(conn *websocket.Conn) {
	for {
		_, b, err := conn.Read(context.Background())
		if err != nil {
			c.mu.Lock()
			if c.conn == conn {
				c.conn = nil
			}
			c.mu.Unlock()
			c.pending.Range(func(k, v any) bool {
				v.(chan response) <- response{ID: k.(int64), Error: json.RawMessage(`{"json":{"message":"connection closed"}}`)}
				c.pending.Delete(k)
				return true
			})
			return
		}
		if string(b) == "PONG" || string(b) == "PING" {
			continue
		}
		var msgs []response
		if len(b) > 0 && b[0] == '[' {
			_ = json.Unmarshal(b, &msgs)
		} else {
			var r response
			if json.Unmarshal(b, &r) == nil {
				msgs = []response{r}
			}
		}
		for _, r := range msgs {
			if ch, ok := c.pending.LoadAndDelete(r.ID); ok {
				ch.(chan response) <- r
			}
		}
	}
}

func (c *Client) call(ctx context.Context, method, path string, input any) (json.RawMessage, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return nil, errors.New("not connected")
	}
	id := c.nextID.Add(1)
	ch := make(chan response, 1)
	c.pending.Store(id, ch)
	msg := map[string]any{"id": id, "method": method, "params": map[string]any{"path": path, "input": map[string]any{"json": input}}}
	b, _ := json.Marshal(msg)
	if err := conn.Write(ctx, websocket.MessageText, b); err != nil {
		c.pending.Delete(id)
		return nil, err
	}
	select {
	case <-ctx.Done():
		c.pending.Delete(id)
		return nil, ctx.Err()
	case r := <-ch:
		if len(r.Error) > 0 {
			var e struct {
				JSON struct {
					Message string `json:"message"`
					Data    struct {
						Code string `json:"code"`
					} `json:"data"`
				} `json:"json"`
			}
			_ = json.Unmarshal(r.Error, &e)
			return nil, &Error{Path: path, Message: e.JSON.Message, Code: e.JSON.Data.Code}
		}
		if r.Result == nil {
			return nil, nil
		}
		var env struct {
			JSON json.RawMessage `json:"json"`
		}
		if err := json.Unmarshal(r.Result.Data, &env); err != nil {
			return nil, err
		}
		return env.JSON, nil
	}
}

func (c *Client) do(ctx context.Context, method, path string, input any) (json.RawMessage, error) {
	for attempt := 0; ; attempt++ {
		if err := c.ensure(ctx); err != nil {
			return nil, err
		}
		out, err := c.call(ctx, method, path, input)
		var te *Error
		if attempt == 0 && (errors.As(err, &te) && te.Code == "UNAUTHORIZED" || err != nil && err.Error() == "not connected") {
			c.reset() // session expired or socket dropped: log in again
			continue
		}
		return out, err
	}
}

func (c *Client) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.CloseNow()
		c.conn = nil
	}
}

// Query runs a tRPC query.
func (c *Client) Query(ctx context.Context, path string, input any) (json.RawMessage, error) {
	return c.do(ctx, "query", path, input)
}

// Mutate runs a tRPC mutation.
func (c *Client) Mutate(ctx context.Context, path string, input any) (json.RawMessage, error) {
	return c.do(ctx, "mutation", path, input)
}
