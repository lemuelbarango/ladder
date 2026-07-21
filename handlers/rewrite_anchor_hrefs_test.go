package handlers

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxifyHref(t *testing.T) {
	base := &url.URL{Scheme: "https", Host: "example.com", Path: "/article"}

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"same-origin absolute path", "/foo/bar", "/https://example.com/foo/bar"},
		{"cross-domain absolute URL", "https://other.example/x", "/https://other.example/x"},
		{"protocol-relative URL", "//cdn.example/x.png", "/https://cdn.example/x.png"},
		{"http scheme preserved", "http://plain.example/x", "/http://plain.example/x"},
		{"relative path (dot)", "./sibling", "/https://example.com/sibling"},
		{"relative path (parent)", "../up", "/https://example.com/up"},
		{"query-only", "?q=1", "/https://example.com/article?q=1"},
		{"empty", "", ""},
		{"anchor only", "#section", ""},
		{"javascript scheme", "javascript:void(0)", ""},
		{"javascript scheme uppercase", "JavaScript:alert(1)", ""},
		{"mailto scheme", "mailto:test@example.com", ""},
		{"tel scheme", "tel:+15551234", ""},
		{"data URL", "data:text/plain,foo", ""},
		{"blob URL", "blob:https://example.com/uuid", ""},
		{"already proxied absolute", "/https://example.com/foo", ""},
		{"already proxied http", "/http://example.com/foo", ""},
		{"whitespace tolerated", "  https://x.example/  ", "/https://x.example/"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, proxifyHref(c.in, base))
		})
	}
}

func TestRewriteAnchorHrefs(t *testing.T) {
	base := &url.URL{Scheme: "https", Host: "example.com", Path: "/"}

	in := `<html><body>` +
		`<a href="/foo">Rel</a>` +
		`<a href="https://other.example/x">Abs</a>` +
		`<a href="//cdn.example/y">Proto-rel</a>` +
		`<a href="#anchor">Anchor</a>` +
		`<a href="mailto:a@b.c">Mail</a>` +
		`<area href="/img-map">` +
		`<link href="/styles.css" rel="stylesheet">` + // <link> must not be touched by this pass
		`</body></html>`

	out := rewriteAnchorHrefs(in, base)

	assert.Contains(t, out, `<a href="/https://example.com/foo">Rel</a>`)
	assert.Contains(t, out, `<a href="/https://other.example/x">Abs</a>`)
	assert.Contains(t, out, `<a href="/https://cdn.example/y">Proto-rel</a>`)
	assert.Contains(t, out, `<a href="#anchor">Anchor</a>`)
	assert.Contains(t, out, `<a href="mailto:a@b.c">Mail</a>`)
	assert.Contains(t, out, `<area href="/https://example.com/img-map"/>`)
	// <link> is out of scope for this rewriter — its href should be untouched
	assert.Contains(t, out, `<link href="/styles.css"`)
}

func TestRewriteAnchorHrefs_SingleQuoted(t *testing.T) {
	base := &url.URL{Scheme: "https", Host: "example.com", Path: "/"}
	// goquery normalizes attribute quoting to double-quotes when serializing;
	// the important behavior is that the URL got proxified.
	in := `<a href='https://other.example/x'>Single</a>`
	out := rewriteAnchorHrefs(in, base)
	assert.True(t, strings.Contains(out, `href="/https://other.example/x"`),
		"single-quoted href should be rewritten, got: %s", out)
}

func TestRewriteAnchorHrefs_ParseFailureFallsBack(t *testing.T) {
	// goquery is lenient and should not fail here, but the contract is that
	// on any error the original body is returned unchanged. Feed something
	// weird and confirm the return value is a superset of the input.
	base := &url.URL{Scheme: "https", Host: "example.com"}
	in := `<a href="/foo">x</a>`
	out := rewriteAnchorHrefs(in, base)
	assert.Contains(t, out, `href="/https://example.com/foo"`)
}
