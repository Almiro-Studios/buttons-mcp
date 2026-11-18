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

func TestSortReadingOrder(t *testing.T) {
	doc := Doc{"type": "section", "items": []any{}}
	for _, p := range [][2]int{{0, 0}, {12, 0}, {0, 1}, {15, 1}, {3, 1}, {11, 0}} {
		Upsert(doc, p[0], p[1])
	}
	floating := Doc{"type": "button", "location": map[string]any{"size": map[string]any{"w": float64(1), "h": float64(1)}}}
	items := doc["items"].([]any)
	doc["items"] = append(items[:2:2], append([]any{floating}, items[2:]...)...)

	SortReadingOrder(doc)

	var got [][2]int
	for i, it := range doc["items"].([]any) {
		m := it.(map[string]any)
		if i == 2 {
			if m["location"].(map[string]any)["staticPosition"] != nil {
				t.Fatalf("floating item moved: slot 2 = %v", m)
			}
			continue
		}
		x, y, _ := Pos(m)
		got = append(got, [2]int{x, y})
	}
	want := [][2]int{{0, 0}, {11, 0}, {12, 0}, {0, 1}, {3, 1}, {15, 1}}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}
