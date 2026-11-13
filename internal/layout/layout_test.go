// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

package layout

import "testing"

func TestUpsertFindDelete(t *testing.T) {
	doc := Doc{"type": "section", "items": []any{}}
	b := Upsert(doc, 3, 1)
	Content(b)["text"] = "hi"
	if got := Find(doc, 3, 1); got == nil || Content(got)["text"] != "hi" {
		t.Fatalf("Find after Upsert = %v", got)
	}
	if Upsert(doc, 3, 1)["id"] != b["id"] {
		t.Fatal("Upsert created a duplicate")
	}
	if !Delete(doc, 3, 1) || Find(doc, 3, 1) != nil {
		t.Fatal("Delete failed")
	}
}

func TestSetColorRules(t *testing.T) {
	b := NewButton(0, 0)
	SetColorRules(b, []ColorRule{{Expression: "$(a.b) > 1", Color: Color{"red", "500"}}})
	fbs := b["feedbacks"].([]any)
	fb := fbs[0].(map[string]any)
	val := fb["values"].(map[string]any)["expression"].(map[string]any)["val"]
	if val != "_EXP_($(a.b) > 1)" {
		t.Fatalf("expression = %v", val)
	}
}
