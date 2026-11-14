// Copyright 2026 Almiro Studios
// SPDX-License-Identifier: MIT

// Package img fetches images for key faces and converts them to the base64
// PNG ("png64") that Buttons stores on a key.
package img

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// size is the edge length SVG icons are rendered at.
const size = 144

// maxBytes caps downloads; key faces are tiny.
const maxBytes = 2 << 20

var linkRe = regexp.MustCompile(`(?is)<link\b[^>]*>`)
var attrRe = regexp.MustCompile(`(?is)(\w[\w-]*)\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)

// SiteIcon finds the best icon a website declares and returns it as base64
// PNG. It prefers apple-touch-icon, then the largest PNG icon, then an SVG
// icon (rendered at 144 px), then common paths, then Google's favicon service.
func SiteIcon(ctx context.Context, site string) (string, error) {
	base, err := url.Parse(site)
	if err != nil || base.Host == "" {
		base, err = url.Parse("https://" + site)
		if err != nil || base.Host == "" {
			return "", fmt.Errorf("not a site URL: %q", site)
		}
	}
	var candidates []string
	if page, err := fetch(ctx, base.String()); err == nil {
		type icon struct {
			href string
			rank int
		}
		var icons []icon
		for _, tag := range linkRe.FindAllString(string(page), -1) {
			attrs := map[string]string{}
			for _, m := range attrRe.FindAllStringSubmatch(tag, -1) {
				attrs[strings.ToLower(m[1])] = strings.Trim(m[2], `"'`)
			}
			rel := strings.ToLower(attrs["rel"])
			href := attrs["href"]
			if href == "" || !strings.Contains(rel, "icon") || strings.Contains(rel, "mask") {
				continue
			}
			u, err := base.Parse(href)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				continue
			}
			rank := 0
			switch {
			case strings.Contains(rel, "apple-touch"):
				rank = 1000
			case strings.Contains(strings.ToLower(attrs["type"]), "svg") || strings.HasSuffix(strings.ToLower(u.Path), ".svg"):
				rank = 500
			default:
				if n, err := strconv.Atoi(strings.SplitN(strings.ToLower(attrs["sizes"]), "x", 2)[0]); err == nil {
					rank = n
				} else {
					rank = 16
				}
			}
			icons = append(icons, icon{u.String(), rank})
		}
		sort.SliceStable(icons, func(i, j int) bool { return icons[i].rank > icons[j].rank })
		for _, ic := range icons {
			candidates = append(candidates, ic.href)
		}
	}
	for _, p := range []string{"/apple-touch-icon.png", "/favicon.svg", "/favicon.png"} {
		u, _ := base.Parse(p)
		candidates = append(candidates, u.String())
	}
	candidates = append(candidates, "https://www.google.com/s2/favicons?sz=128&domain_url="+url.QueryEscape(base.Scheme+"://"+base.Host))
	var lastErr error
	for _, c := range candidates {
		if s, err := PNG64(ctx, c); err == nil {
			return s, nil
		} else {
			lastErr = err
		}
	}
	return "", fmt.Errorf("no usable icon for %s: %w", base.Host, lastErr)
}

// File reads a local PNG, JPEG, GIF or SVG file and returns base64 PNG.
func File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) > maxBytes {
		return "", fmt.Errorf("image larger than %d bytes", maxBytes)
	}
	return Encode(data)
}

func fetch(ctx context.Context, src string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "buttons-mcp")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: HTTP %d", src, res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxBytes {
		return nil, fmt.Errorf("%s is larger than %d bytes", src, maxBytes)
	}
	return data, nil
}

// PNG64 downloads a PNG, JPEG, GIF or SVG image and returns it as base64 PNG.
func PNG64(ctx context.Context, src string) (string, error) {
	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		return "", fmt.Errorf("image_url must be http(s)")
	}
	data, err := fetch(ctx, src)
	if err != nil {
		return "", err
	}
	return Encode(data)
}

// Encode converts PNG, JPEG, GIF or SVG bytes to base64 PNG.
func Encode(data []byte) (string, error) {
	m, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if m, err = renderSVG(data); err != nil {
			return "", fmt.Errorf("unsupported image (use PNG, JPEG, GIF or SVG): %w", err)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func renderSVG(data []byte) (image.Image, error) {
	if !bytes.Contains(data[:min(len(data), 1024)], []byte("<svg")) {
		return nil, fmt.Errorf("not an SVG")
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data), oksvg.IgnoreErrorMode)
	if err != nil {
		return nil, err
	}
	icon.SetTarget(0, 0, size, size)
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	icon.Draw(rasterx.NewDasher(size, size, rasterx.NewScannerGV(size, size, rgba, rgba.Bounds())), 1)
	return rgba, nil
}
