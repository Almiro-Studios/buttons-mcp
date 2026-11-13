// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

// Command buttons-mcp is an MCP server for Bitfocus Buttons. It talks to the
// same tRPC-over-WebSocket API the Buttons web UI uses.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Almiro-Studios/buttons-mcp/internal/creds"
	"github.com/Almiro-Studios/buttons-mcp/internal/layout"
	"github.com/Almiro-Studios/buttons-mcp/internal/trpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

// serverURL reads BUTTONS_URL, accepting either the web address
// (https://localhost:4443) or the WebSocket endpoint.
func serverURL() string {
	u := strings.TrimRight(env("BUTTONS_URL", "https://localhost:4443"), "/")
	u = strings.Replace(u, "https://", "wss://", 1)
	u = strings.Replace(u, "http://", "ws://", 1)
	if !strings.HasSuffix(u, "/api/wstrpc") {
		u += "/api/wstrpc"
	}
	return u
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func newClient() *trpc.Client {
	url := serverURL()
	c := &trpc.Client{URL: url, Insecure: env("BUTTONS_INSECURE_TLS", "true") == "true"}
	if l, err := creds.Load(url); err == nil {
		c.Username, c.Password = l.Username, l.Password
	} else {
		log.Print(err)
	}
	return c
}

func login(url string) {
	l, err := creds.Prompt()
	if err != nil {
		log.Fatal(err)
	}
	c := &trpc.Client{URL: url, Username: l.Username, Password: l.Password, Insecure: env("BUTTONS_INSECURE_TLS", "true") == "true"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := c.Query(ctx, "position.all", nil); err != nil {
		log.Fatalf("login failed: %v", err)
	}
	if err := creds.Save(url, l); err != nil {
		log.Fatalf("save to keychain: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Logged in as %s; saved to the system keychain for %s\n", l.Username, url)
}

type saved struct {
	Revision string     `json:"revision"`
	Document layout.Doc `json:"document"`
}

func load(ctx context.Context, c *trpc.Client, positionID string) (*saved, error) {
	raw, err := c.Query(ctx, "positionDocument.saved", map[string]any{"resourceId": positionID})
	if err != nil {
		return nil, err
	}
	var s saved
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("decode layout: %w", err)
	}
	if s.Document == nil {
		return nil, fmt.Errorf("position %s has no saved layout", positionID)
	}
	return &s, nil
}

func save(ctx context.Context, c *trpc.Client, positionID string, s *saved) error {
	_, err := c.Mutate(ctx, "positionDocument.save", map[string]any{
		"resourceId":       positionID,
		"expectedRevision": s.Revision,
		"document":         s.Document,
	})
	return err
}

// edit loads a layout, applies f, and saves it.
func edit(ctx context.Context, c *trpc.Client, positionID string, f func(doc layout.Doc) error) error {
	s, err := load(ctx, c, positionID)
	if err != nil {
		return err
	}
	if err := f(s.Document); err != nil {
		return err
	}
	return save(ctx, c, positionID, s)
}

func text(format string, a ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, a...)}}}
}

func jsonResult(v any) *mcp.CallToolResult {
	b, _ := json.MarshalIndent(v, "", "  ")
	return text("%s", b)
}

type empty struct{}

type posArgs struct {
	PositionID string `json:"position_id" jsonschema:"position id from list_positions"`
}

type keyArgs struct {
	PositionID string `json:"position_id" jsonschema:"position id from list_positions"`
	X          int    `json:"x" jsonschema:"column, 0-based from the left"`
	Y          int    `json:"y" jsonschema:"row, 0-based from the top"`
}

type setButtonArgs struct {
	keyArgs
	Text    *string       `json:"text,omitempty" jsonschema:"key text; variables as $(connectionKey.variable), e.g. CI $(FileReader.contents)"`
	Icon    *string       `json:"icon,omitempty" jsonschema:"icon name, e.g. shield; empty string removes the icon"`
	BgColor *layout.Color `json:"bg_color,omitempty" jsonschema:"base background as [name, shade], e.g. [\"green\",\"500\"]"`
}

type colorRulesArgs struct {
	keyArgs
	Rules []layout.ColorRule `json:"rules" jsonschema:"rules in order; each sets the background when its expression is true"`
}

type gradientArgs struct {
	keyArgs
	Variable string  `json:"variable" jsonschema:"numeric variable, e.g. FileReader.contents"`
	Low      float64 `json:"low" jsonschema:"value at or below which the key is fully green"`
	High     float64 `json:"high" jsonschema:"value at or above which the key is fully red"`
	Steps    int     `json:"steps,omitempty" jsonschema:"number of colour bands, 2 to 8 (default 5)"`
}

var gradientPalette = []layout.Color{
	{"green", "600"}, {"lime", "500"}, {"yellow", "500"}, {"amber", "500"},
	{"orange", "500"}, {"orange", "600"}, {"red", "500"}, {"red", "700"},
}

func pickColors(n int) []layout.Color {
	out := make([]layout.Color, n)
	for i := range out {
		out[i] = gradientPalette[i*(len(gradientPalette)-1)/(n-1)]
	}
	return out
}

type callArgs struct {
	Kind  string `json:"kind" jsonschema:"query (read) or mutation (change)"`
	Path  string `json:"path" jsonschema:"tRPC procedure, e.g. connection.list, position.create, positionDocument.save"`
	Input any    `json:"input,omitempty" jsonschema:"procedure input: a JSON object, array or value (omit for none)"`
}

type listActionsArgs struct {
	Connection string `json:"connection,omitempty" jsonschema:"only this connection (label, id or module, case-insensitive substring)"`
}

type setActionsArgs struct {
	keyArgs
	Press   []layout.Action `json:"press,omitempty" jsonschema:"actions run when the key goes down, in order; omit to leave unchanged, [] to clear"`
	Release []layout.Action `json:"release,omitempty" jsonschema:"actions run when the key comes up; omit to leave unchanged, [] to clear"`
}

type connActions struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	ModuleID string `json:"moduleId"`
	Actions  []struct {
		ID          string           `json:"id"`
		Label       string           `json:"label"`
		Description string           `json:"description"`
		Options     []map[string]any `json:"options"`
	} `json:"actions"`
}

func allActions(ctx context.Context, c *trpc.Client) ([]connActions, error) {
	raw, err := c.Query(ctx, "connection.allActions", map[string]any{})
	if err != nil {
		return nil, err
	}
	var out []connActions
	return out, json.Unmarshal(raw, &out)
}

// resolveActions turns tool input into stored actions, looking up each
// action's connection and module.
func resolveActions(conns []connActions, in []layout.Action) ([]any, error) {
	if in == nil {
		return nil, nil
	}
	out := []any{}
	for _, a := range in {
		connID, _, ok := strings.Cut(a.ActionID, ":")
		if !ok {
			return nil, fmt.Errorf("action_id %q should look like <connectionId>:<action>; see list_actions", a.ActionID)
		}
		var mod string
		found := false
		for _, cn := range conns {
			if cn.ID != connID {
				continue
			}
			mod = cn.ModuleID
			for _, act := range cn.Actions {
				if act.ID == a.ActionID {
					found = true
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown action %q; see list_actions", a.ActionID)
		}
		out = append(out, layout.ConnectionAction(a, connID, mod))
	}
	return out, nil
}

func register(s *mcp.Server, c *trpc.Client) {
	mcp.AddTool(s, &mcp.Tool{Name: "list_actions", Description: "List the actions each connection offers (switch ATEM input, route Videohub, start stream, …) with their option ids, types, defaults and choices. Use the ids with set_button_actions."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a listActionsArgs) (*mcp.CallToolResult, any, error) {
			conns, err := allActions(ctx, c)
			if err != nil {
				return nil, nil, err
			}
			f := strings.ToLower(a.Connection)
			var out []map[string]any
			for _, cn := range conns {
				if f != "" && !strings.Contains(strings.ToLower(cn.ID+" "+cn.Label+" "+cn.ModuleID), f) {
					continue
				}
				var acts []map[string]any
				for _, act := range cn.Actions {
					var opts []map[string]any
					for _, o := range act.Options {
						opt := map[string]any{"id": o["id"], "label": o["label"], "type": o["type"], "default": o["default"]}
						if ch, ok := o["choices"].([]any); ok {
							var ids []any
							for _, x := range ch {
								if m, ok := x.(map[string]any); ok {
									ids = append(ids, m["id"])
								}
							}
							opt["choices"] = ids
						}
						opts = append(opts, opt)
					}
					acts = append(acts, map[string]any{"action_id": act.ID, "label": act.Label, "options": opts})
				}
				out = append(out, map[string]any{"connection": cn.Label, "connection_id": cn.ID, "module": cn.ModuleID, "actions": acts})
			}
			return jsonResult(out), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "set_button_actions", Description: "Set what the key at x,y does when pressed and released: a list of connection actions (from list_actions) with option values and optional delays. Creates the key if needed. Saves and applies the layout."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a setActionsArgs) (*mcp.CallToolResult, any, error) {
			conns, err := allActions(ctx, c)
			if err != nil {
				return nil, nil, err
			}
			press, err := resolveActions(conns, a.Press)
			if err != nil {
				return nil, nil, err
			}
			release, err := resolveActions(conns, a.Release)
			if err != nil {
				return nil, nil, err
			}
			err = edit(ctx, c, a.PositionID, func(doc layout.Doc) error {
				layout.SetActions(layout.Upsert(doc, a.X, a.Y), press, release)
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			return text("key %d,%d: %d press and %d release actions", a.X, a.Y, len(press), len(release)), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{
		Name: "call_api",
		Description: "Call any Buttons API procedure directly (the internal tRPC API the web UI uses). " +
			"Use for anything the dedicated tools do not cover: connections (connection.list, connection.get, connection.add), " +
			"positions (position.all, position.get, position.create, position.update, position.delete), surfaces, presets and more. " +
			"Mutations change the live system; read the current state with a query first.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a callArgs) (*mcp.CallToolResult, any, error) {
		input := a.Input
		var raw json.RawMessage
		var err error
		switch a.Kind {
		case "query":
			raw, err = c.Query(ctx, a.Path, input)
		case "mutation":
			raw, err = c.Mutate(ctx, a.Path, input)
		default:
			return nil, nil, fmt.Errorf("kind must be query or mutation")
		}
		if err != nil {
			return nil, nil, err
		}
		if len(raw) == 0 {
			raw = json.RawMessage("null")
		}
		return text("%s", raw), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "list_positions", Description: "List Buttons positions (layouts mapped to surfaces such as a Stream Deck)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
			raw, err := c.Query(ctx, "position.all", nil)
			if err != nil {
				return nil, nil, err
			}
			var all []map[string]any
			_ = json.Unmarshal(raw, &all)
			var out []map[string]any
			for _, p := range all {
				out = append(out, map[string]any{"id": p["id"], "label": p["label"], "description": p["description"]})
			}
			return jsonResult(out), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "list_buttons", Description: "List the keys on a position with their grid location, text, background and colour-rule count."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a posArgs) (*mcp.CallToolResult, any, error) {
			sv, err := load(ctx, c, a.PositionID)
			if err != nil {
				return nil, nil, err
			}
			var out []map[string]any
			for _, b := range layout.Buttons(sv.Document) {
				x, y, _ := layout.Pos(b)
				ct := layout.Content(b)
				fbs, _ := b["feedbacks"].([]any)
				out = append(out, map[string]any{"x": x, "y": y, "text": ct["text"], "icon": ct["icon"], "bg_color": ct["bgColor"], "color_rules": len(fbs)})
			}
			return jsonResult(out), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "get_layout", Description: "Return the raw saved layout document and revision for a position (for advanced edits)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a posArgs) (*mcp.CallToolResult, any, error) {
			sv, err := load(ctx, c, a.PositionID)
			if err != nil {
				return nil, nil, err
			}
			return jsonResult(sv), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "set_button", Description: "Create or update the key at x,y: text (with $(...) variables), icon and base background colour. Saves and applies the layout."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a setButtonArgs) (*mcp.CallToolResult, any, error) {
			err := edit(ctx, c, a.PositionID, func(doc layout.Doc) error {
				b := layout.Upsert(doc, a.X, a.Y)
				ct := layout.Content(b)
				if a.Text != nil {
					ct["text"] = *a.Text
				}
				if a.Icon != nil {
					if *a.Icon == "" {
						delete(ct, "icon")
					} else {
						ct["icon"] = *a.Icon
					}
				}
				if a.BgColor != nil {
					ct["bgColor"] = []any{a.BgColor[0], a.BgColor[1]}
				}
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			return text("saved key %d,%d", a.X, a.Y), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "delete_button", Description: "Remove the key at x,y and save the layout."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a keyArgs) (*mcp.CallToolResult, any, error) {
			found := false
			err := edit(ctx, c, a.PositionID, func(doc layout.Doc) error {
				found = layout.Delete(doc, a.X, a.Y)
				if !found {
					return fmt.Errorf("no key at %d,%d", a.X, a.Y)
				}
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			return text("deleted key %d,%d", a.X, a.Y), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "set_color_rules", Description: "Replace a key's colour rules: each rule sets the background when its Buttons expression is true (e.g. $(FileReader.contents) >= 15). Empty list removes all rules."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a colorRulesArgs) (*mcp.CallToolResult, any, error) {
			err := edit(ctx, c, a.PositionID, func(doc layout.Doc) error {
				b := layout.Find(doc, a.X, a.Y)
				if b == nil {
					return fmt.Errorf("no key at %d,%d", a.X, a.Y)
				}
				if len(a.Rules) == 0 {
					b["feedbacks"] = []any{}
					return nil
				}
				layout.SetColorRules(b, a.Rules)
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			return text("set %d colour rules on key %d,%d", len(a.Rules), a.X, a.Y), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "set_load_gradient", Description: "Colour a key green→red by a numeric variable: splits low..high into bands (Buttons has a fixed palette, so this is a stepped gradient)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, a gradientArgs) (*mcp.CallToolResult, any, error) {
			n := a.Steps
			if n == 0 {
				n = 5
			}
			if n < 2 || n > 8 || a.High <= a.Low {
				return nil, nil, fmt.Errorf("need 2-8 steps and high > low")
			}
			colors := pickColors(n)
			v := "$(" + strings.TrimSuffix(strings.TrimPrefix(a.Variable, "$("), ")") + ")"
			step := (a.High - a.Low) / float64(n-1)
			var rules []layout.ColorRule
			for i := 1; i < n; i++ {
				lo := a.Low + step*float64(i-1)
				expr := fmt.Sprintf("%s > %g", v, lo)
				if i < n-1 {
					expr += fmt.Sprintf(" && %s <= %g", v, lo+step)
				}
				rules = append(rules, layout.ColorRule{Expression: expr, Color: colors[i]})
			}
			err := edit(ctx, c, a.PositionID, func(doc layout.Doc) error {
				b := layout.Find(doc, a.X, a.Y)
				if b == nil {
					return fmt.Errorf("no key at %d,%d", a.X, a.Y)
				}
				layout.Content(b)["bgColor"] = []any{colors[0][0], colors[0][1]}
				layout.SetColorRules(b, rules)
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
			return jsonResult(map[string]any{"base": colors[0], "rules": rules}), nil, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "list_variables", Description: "List connection variables with current values, for use as $(key.variable) in key text and rules."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
			raw, err := c.Query(ctx, "connection.list", map[string]any{})
			if err != nil {
				return nil, nil, err
			}
			return text("%s", raw), nil, nil
		})
}

func main() {
	url := serverURL()
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "login":
			login(url)
			return
		case "logout":
			if err := creds.Delete(url); err != nil {
				log.Fatal(err)
			}
			fmt.Fprintln(os.Stderr, "Removed saved login for", url)
			return
		case "version":
			fmt.Println(version)
			return
		}
	}
	c := newClient()
	if len(os.Args) > 2 && os.Args[1] == "dump" { // debugging aid: buttons-mcp dump <positionId|->
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var raw json.RawMessage
		var err error
		if os.Args[2] == "-" {
			raw, err = c.Query(ctx, "position.all", nil)
		} else {
			raw, err = c.Query(ctx, "positionDocument.saved", map[string]any{"resourceId": os.Args[2]})
		}
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write(raw)
		return
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "buttons-mcp", Version: version}, nil)
	register(s, c)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
