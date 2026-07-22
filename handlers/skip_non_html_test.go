package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHTMLResponse(t *testing.T) {
	mk := func(ct string) *http.Response {
		h := http.Header{}
		if ct != "" {
			h.Set("Content-Type", ct)
		}
		return &http.Response{Header: h}
	}

	cases := []struct {
		name string
		ct   string
		want bool
	}{
		{"plain html", "text/html", true},
		{"html with charset", "text/html; charset=utf-8", true},
		{"html with weird spacing", "  Text/HTML ; charset=utf-8", true},
		{"xhtml", "application/xhtml+xml", true},
		{"missing content-type", "", true}, // permissive fallback

		{"xml sitemap", "application/xml", false},
		{"text xml", "text/xml", false},
		{"rss", "application/rss+xml", false},
		{"atom", "application/atom+xml", false},
		{"json", "application/json", false},
		{"json with params", "application/json; charset=utf-8", false},
		{"css", "text/css", false},
		{"javascript", "text/javascript", false},
		{"application javascript", "application/javascript", false},
		{"png", "image/png", false},
		{"pdf", "application/pdf", false},
		{"plain text", "text/plain", false},
		{"octet-stream", "application/octet-stream", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, isHTMLResponse(mk(c.ct)))
		})
	}
}

func TestIsHTMLResponse_NilResponse(t *testing.T) {
	// Nil response falls back to permissive/HTML — mirrors previous behavior
	// when no response was available.
	assert.True(t, isHTMLResponse(nil))
}
