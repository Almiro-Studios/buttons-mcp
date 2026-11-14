// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

package img

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"
)

func TestEncodeRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	s, err := Encode(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := base64.StdEncoding.DecodeString(s)
	if _, err := png.Decode(bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
	if _, err := Encode([]byte("nope")); err == nil {
		t.Fatal("expected error for non-image")
	}
}

func TestEncodeSVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10" fill="#f00"/></svg>`)
	s, err := Encode(svg)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := base64.StdEncoding.DecodeString(s)
	m, err := png.Decode(bytes.NewReader(b))
	if err != nil || m.Bounds().Dx() != size {
		t.Fatalf("got %v, %v", m.Bounds(), err)
	}
}

func TestEncodeSVGKeepsAspect(t *testing.T) {
	// A wide red bar must not be stretched to fill the square.
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 10"><rect width="20" height="10" fill="#f00"/></svg>`)
	s, err := Encode(svg)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := base64.StdEncoding.DecodeString(s)
	m, _ := png.Decode(bytes.NewReader(b))
	if _, _, _, a := m.At(size/2, 2).RGBA(); a != 0 {
		t.Fatal("top edge should be transparent for a 2:1 image")
	}
	if r, _, _, _ := m.At(size/2, size/2).RGBA(); r == 0 {
		t.Fatal("centre should be red")
	}
}
