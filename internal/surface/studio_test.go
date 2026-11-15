// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

package surface

import "testing"

func TestGroups(t *testing.T) {
	// Each group is 8 wide and 2 tall. Group 1 positions 9-11 sit under 1-3.
	// Group 2 top positions 1-3 are the cells we must not use for those.
	for _, tc := range []struct{ key, x, y, group, index int }{
		{1, 0, 0, 1, 1},
		{2, 1, 0, 1, 2},
		{3, 2, 0, 1, 3},
		{8, 7, 0, 1, 8},
		{9, 0, 1, 1, 9},
		{10, 1, 1, 1, 10},
		{11, 2, 1, 1, 11},
		{16, 7, 1, 1, 16},
		{17, 8, 0, 2, 1},
		{18, 9, 0, 2, 2},
		{19, 10, 0, 2, 3},
		{24, 15, 0, 2, 8},
		{25, 8, 1, 2, 9},
		{32, 15, 1, 2, 16},
	} {
		x, y, ok := XY(tc.key)
		if !ok || x != tc.x || y != tc.y {
			t.Fatalf("XY(%d) = %d,%d ok=%v, want %d,%d", tc.key, x, y, ok, tc.x, tc.y)
		}
		if got, ok := Key(tc.x, tc.y); !ok || got != tc.key {
			t.Fatalf("Key(%d,%d) = %d, want %d", tc.x, tc.y, got, tc.key)
		}
		g, i, ok := Group(tc.key)
		if !ok || g != tc.group || i != tc.index {
			t.Fatalf("Group(%d) = %d,%d, want %d,%d", tc.key, g, i, tc.group, tc.index)
		}
	}
	if _, _, ok := XY(0); ok {
		t.Fatal("key 0 should be out of range")
	}
	if _, _, ok := Group(33); ok {
		t.Fatal("key 33 should be out of range")
	}
	where, ok := Where(0, 1)
	if !ok || where != "left group position 9, directly under left position 1" {
		t.Fatalf("Where(0,1) = %q", where)
	}
	where, ok = Where(8, 0)
	if !ok || where != "right group position 1, top row" {
		t.Fatalf("Where(8,0) = %q", where)
	}
}
