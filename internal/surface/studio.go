// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

// Package surface describes the Stream Deck Studio grid Buttons uses, so a
// caller can turn "group 1, position 9" into layout coordinates.
package surface

import "fmt"

// Stream Deck Studio (model StreamDeckStudio) is two side-by-side groups of
// 16 keys. Each group is 8 columns by 2 rows, numbered left to right, then
// the row below.
//
//	Group 1 (left, x=0–7):  positions 1–8 on y=0, positions 9–16 on y=1.
//	Group 2 (right, x=8–15): positions 1–8 on y=0, positions 9–16 on y=1.
//
// Position 9 of group 1 is x=0,y=1, directly under position 1. Position 1
// of group 2 is x=8,y=0. The left encoder shares group 1 position 1
// (0,0); the right encoder shares group 2 position 8 (15,0).
const (
	Columns      = 16
	GroupColumns = 8
	Rows         = 2
	Keys         = Columns * Rows
	GroupSize    = GroupColumns * Rows
)

// Instructions is the MCP server hint. Small models read this once, then
// call describe_surface for the live map.
const Instructions = "Stream Deck Studio is two side-by-side groups of 16. Each group is 8 wide and 2 tall, numbered left to right, then the row below. " +
	"Group 1 is the left half (x=0-7): positions 1-3 are x=0,1,2 y=0; positions 9-11 are x=0,1,2 y=1, directly under them. " +
	"Group 2 is the right half (x=8-15): its top positions 1-3 are x=8,9,10 y=0. Do not put group 1 positions 9-11 there. " +
	"Keys 1-16 are group 1 and keys 17-32 are group 2, in that same order. " +
	"Left encoder shares key 1; right encoder shares key 24 (group 2 position 8). " +
	"position_id is the layout, not a key. Pass key (1-32) or x and y. Call describe_surface before editing a layout."

// XY converts a 1-based key number into 0-based column and row.
// Keys 1–16 fill group 1 (8×2, left). Keys 17–32 fill group 2 (right).
func XY(key int) (x, y int, ok bool) {
	if key < 1 || key > Keys {
		return 0, 0, false
	}
	i := key - 1
	group := i / GroupSize
	local := i % GroupSize
	return group*GroupColumns + local%GroupColumns, local / GroupColumns, true
}

// Key converts a column and row into a 1-based key number.
func Key(x, y int) (int, bool) {
	if x < 0 || x >= Columns || y < 0 || y >= Rows {
		return 0, false
	}
	group := x / GroupColumns
	local := y*GroupColumns + x%GroupColumns
	return group*GroupSize + local + 1, true
}

// Group reports which side a key is on. Group 1 is the left 16 (keys 1–16);
// group 2 is the right 16 (keys 17–32). Index is the 1-based position
// inside that group.
func Group(key int) (group, index int, ok bool) {
	if key < 1 || key > Keys {
		return 0, 0, false
	}
	return (key-1)/GroupSize + 1, (key-1)%GroupSize + 1, true
}

// Where is a one-line description of a cell, in the two-group numbering.
// The cell under left position 1 is "left group position 9, directly under left position 1".
func Where(x, y int) (string, bool) {
	key, ok := Key(x, y)
	if !ok {
		return "", false
	}
	group, index, _ := Group(key)
	side := "left"
	if group == 2 {
		side = "right"
	}
	if index > GroupColumns {
		return fmt.Sprintf("%s group position %d, directly under %s position %d", side, index, side, index-GroupColumns), true
	}
	return fmt.Sprintf("%s group position %d, top row", side, index), true
}

// Summary is the static cheat sheet describe_surface returns.
func Summary() map[string]any {
	return map[string]any{
		"model":       "StreamDeckStudio",
		"shape":       "Two side-by-side groups of 16. Each group is 8 columns by 2 rows.",
		"grid":        "LEFT  1 2 3 4 5 6 7 8 / 9 10 11 12 13 14 15 16     RIGHT  1 2 3 4 5 6 7 8 / 9 10 11 12 13 14 15 16",
		"group_1":     "Left half, keys 1-16, x=0-7. Positions 1-8 are y=0 (x=0-7). Positions 9-16 are y=1 (x=0-7). Position 9 is x=0,y=1, directly under position 1. Positions 10 and 11 are x=1,y=1 and x=2,y=1.",
		"group_2":     "Right half, keys 17-32, x=8-15. Its top positions 1-3 are x=8,9,10 y=0. Do not use these for group 1 positions 9-11.",
		"rule":        "Number each group of 16 on its own: across the top 8, then across the bottom 8. Group 1 position 9 is under group 1 position 1, not at the start of group 2.",
		"encoders":    []any{"left shares key 1 / group 1 position 1 (x=0,y=0)", "right shares key 24 / group 2 position 8 (x=15,y=0)"},
		"place_with":  "pass key 9, 10, 11 for the three cells under the first three keys of group 1",
		"position_id": "position_id is the layout id from list_positions, not a key number",
	}
}
