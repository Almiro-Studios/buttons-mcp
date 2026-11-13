// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

// Package layout edits Buttons position documents (the JSON tree the
// position editor saves). Documents are kept as generic maps so fields this
// package does not know about survive a load/save round trip.
package layout

import (
	"crypto/rand"
)

// Doc is a JSON object from a layout document.
type Doc = map[string]any

// Color is a palette entry such as ["green","500"].
type Color [2]string

// Buttons returns every button node in the document tree.
func Buttons(doc Doc) []Doc {
	var out []Doc
	var walk func(n Doc)
	walk = func(n Doc) {
		if n["type"] == "button" {
			out = append(out, n)
		}
		items, _ := n["items"].([]any)
		for _, it := range items {
			if m, ok := it.(map[string]any); ok {
				walk(m)
			}
		}
	}
	walk(doc)
	return out
}

// Pos returns a button's static grid position.
func Pos(b Doc) (x, y int, ok bool) {
	loc, _ := b["location"].(map[string]any)
	sp, _ := loc["staticPosition"].(map[string]any)
	fx, ok1 := sp["x"].(float64)
	fy, ok2 := sp["y"].(float64)
	return int(fx), int(fy), ok1 && ok2
}

// Find returns the button at x,y in the top-level section, or nil.
func Find(doc Doc, x, y int) Doc {
	for _, b := range Buttons(doc) {
		if bx, by, ok := Pos(b); ok && bx == x && by == y {
			return b
		}
	}
	return nil
}

// newID returns a random 12-character id like the ones the editor generates.
func newID() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// NewButton builds a basic button matching what the editor creates.
func NewButton(x, y int) Doc {
	return Doc{
		"id":            newID(),
		"type":          "button",
		"items":         []any{},
		"buttonType":    "basic",
		"schemaVersion": float64(5),
		"actionSteps": []any{map[string]any{"actionSet": map[string]any{
			"down": map[string]any{"actions": []any{}, "options": map[string]any{}},
			"up":   map[string]any{"actions": []any{}, "options": map[string]any{}},
		}}},
		"feedbacks":     []any{},
		"location":      map[string]any{"size": map[string]any{"w": float64(1), "h": float64(1)}, "staticPosition": map[string]any{"x": float64(x), "y": float64(y)}},
		"buttonStyle":   "iconAndText",
		"buttonContent": map[string]any{"text": ""},
	}
}

// Upsert returns the button at x,y, creating it in the root section if needed.
func Upsert(doc Doc, x, y int) Doc {
	if b := Find(doc, x, y); b != nil {
		return b
	}
	b := NewButton(x, y)
	items, _ := doc["items"].([]any)
	doc["items"] = append(items, b)
	return b
}

// Delete removes the button at x,y. It reports whether one was removed.
func Delete(doc Doc, x, y int) bool {
	var del func(n Doc) bool
	del = func(n Doc) bool {
		items, _ := n["items"].([]any)
		for i, it := range items {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			if m["type"] == "button" {
				if bx, by, ok := Pos(m); ok && bx == x && by == y {
					n["items"] = append(items[:i:i], items[i+1:]...)
					return true
				}
			}
			if del(m) {
				return true
			}
		}
		return false
	}
	return del(doc)
}

// Content returns (and creates) the button's buttonContent map.
func Content(b Doc) map[string]any {
	c, ok := b["buttonContent"].(map[string]any)
	if !ok {
		c = map[string]any{}
		b["buttonContent"] = c
	}
	return c
}

// ColorRule colours a button when Expression evaluates true.
type ColorRule struct {
	Expression string `json:"expression" jsonschema:"Buttons expression, e.g. $(FileReader.contents) >= 15"`
	Color      Color  `json:"color" jsonschema:"palette colour as [name, shade], e.g. [\"red\",\"500\"]"`
}

// ExpressionFeedback builds an Internal > Expression feedback that sets the
// background (and context background) when expr is true.
func ExpressionFeedback(expr string, c Color) map[string]any {
	return map[string]any{
		"id":         newID(),
		"entityType": "feedback",
		"type":       "internal",
		"label":      "Expression",
		"feedbackId": "expression",
		"values": map[string]any{"expression": map[string]any{
			"_adv": true, "type": "exp", "val": "_EXP_(" + expr + ")",
		}},
		"content": map[string]any{"buttonContent": map[string]any{
			"bgColor":        []any{c[0], c[1]},
			"contextBgColor": []any{c[0], c[1]},
		}},
		"enabled": true,
	}
}

// SetColorRules replaces the button's feedbacks with one expression feedback
// per rule. Later rules win when several are true.
func SetColorRules(b Doc, rules []ColorRule) {
	fbs := make([]any, 0, len(rules))
	for _, r := range rules {
		fbs = append(fbs, ExpressionFeedback(r.Expression, r.Color))
	}
	b["feedbacks"] = fbs
}

// Action is one step a key runs on press or release.
type Action struct {
	ActionID string         `json:"action_id" jsonschema:"action id from list_actions, e.g. <connectionId>:readFile"`
	Values   map[string]any `json:"values,omitempty" jsonschema:"option values keyed by option id (see list_actions)"`
	DelayMS  int            `json:"delay_ms,omitempty" jsonschema:"wait this many milliseconds before running"`
}

// ConnectionAction builds a stored connection action. moduleID is the
// connection's module id with version, e.g. bmd-atem:3.14.0.
func ConnectionAction(a Action, connectionID, moduleID string) map[string]any {
	values := a.Values
	if values == nil {
		values = map[string]any{}
	}
	return map[string]any{
		"id":                 newID(),
		"entityType":         "action",
		"type":               "connection",
		"connectionId":       connectionID,
		"connectionModuleId": moduleID,
		"actionId":           a.ActionID,
		"values":             values,
		"delay":              float64(a.DelayMS),
		"enabled":            true,
	}
}

// SetActions replaces the press ("down") and release ("up") actions of the
// key's first step. A nil slice leaves that handler unchanged.
func SetActions(b Doc, press, release []any) {
	steps, _ := b["actionSteps"].([]any)
	if len(steps) == 0 {
		steps = NewButton(0, 0)["actionSteps"].([]any)
		b["actionSteps"] = steps
	}
	step, _ := steps[0].(map[string]any)
	set, _ := step["actionSet"].(map[string]any)
	if set == nil {
		set = map[string]any{}
		step["actionSet"] = set
	}
	for key, acts := range map[string][]any{"down": press, "up": release} {
		if acts == nil {
			continue
		}
		h, _ := set[key].(map[string]any)
		if h == nil {
			h = map[string]any{"options": map[string]any{}}
			set[key] = h
		}
		h["actions"] = acts
	}
}
