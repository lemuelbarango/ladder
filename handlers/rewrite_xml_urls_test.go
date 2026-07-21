package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsXMLResponse(t *testing.T) {
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
		{"application/xml", "application/xml", true},
		{"text/xml", "text/xml", true},
		{"rss", "application/rss+xml", true},
		{"atom", "application/atom+xml", true},
		{"xml with charset", "application/xml; charset=utf-8", true},
		{"custom +xml subtype", "application/xrd+xml", true},
		{"case insensitive", "Application/XML", true},

		{"xhtml is not XML (goes through HTML path)", "application/xhtml+xml", false},
		{"html", "text/html", false},
		{"json", "application/json", false},
		{"plain text", "text/plain", false},
		{"missing", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, isXMLResponse(mk(c.ct)))
		})
	}
	assert.False(t, isXMLResponse(nil))
}

func TestRewriteXMLURLs_Sitemap(t *testing.T) {
	in := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
 <url>
  <loc>https://www.thestar.com/anniversary/</loc>
  <changefreq>hourly</changefreq>
  <priority>0.5000</priority>
 </url>
 <url>
  <loc>https://www.thestar.com/toronto/</loc>
 </url>
</urlset>`
	out := rewriteXMLURLs(in)
	assert.Contains(t, out, `<loc>/https://www.thestar.com/anniversary/</loc>`)
	assert.Contains(t, out, `<loc>/https://www.thestar.com/toronto/</loc>`)
	// PI must be preserved verbatim
	assert.Contains(t, out, `<?xml version="1.0" encoding="UTF-8"?>`)
	// changefreq must not be touched
	assert.Contains(t, out, `<changefreq>hourly</changefreq>`)
}

func TestRewriteXMLURLs_RSS(t *testing.T) {
	in := `<?xml version="1.0"?><rss version="2.0"><channel>
<title>Feed</title>
<link>https://example.com/</link>
<item>
  <title>Post</title>
  <link>https://example.com/post-1</link>
  <guid>https://example.com/post-1</guid>
  <description>See https://example.com/post-2 for more</description>
</item>
</channel></rss>`
	out := rewriteXMLURLs(in)
	assert.Contains(t, out, `<link>/https://example.com/</link>`)
	assert.Contains(t, out, `<link>/https://example.com/post-1</link>`)
	assert.Contains(t, out, `<guid>/https://example.com/post-1</guid>`)
	// URL inside <description> text is NOT rewritten — it's content, not nav
	assert.Contains(t, out, `See https://example.com/post-2 for more`)
	assert.Contains(t, out, `<title>Feed</title>`)
}

func TestRewriteXMLURLs_Atom(t *testing.T) {
	in := `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom">
<title>Feed</title>
<link href="https://example.com/" rel="alternate"/>
<entry>
  <title>Post</title>
  <link href="https://example.com/post-1"/>
</entry>
</feed>`
	out := rewriteXMLURLs(in)
	assert.Contains(t, out, `<link href="/https://example.com/" rel="alternate"/>`)
	assert.Contains(t, out, `<link href="/https://example.com/post-1"/>`)
}

func TestRewriteXMLURLs_LeavesRelativeAlone(t *testing.T) {
	in := `<urlset><url><loc>/relative/path</loc></url></urlset>`
	out := rewriteXMLURLs(in)
	assert.Contains(t, out, `<loc>/relative/path</loc>`)
	assert.Equal(t, in, out)
}

func TestRewriteXMLURLs_Idempotent(t *testing.T) {
	in := `<urlset><url><loc>https://example.com/x</loc></url></urlset>`
	once := rewriteXMLURLs(in)
	twice := rewriteXMLURLs(once)
	assert.Equal(t, once, twice)
	assert.Contains(t, once, `<loc>/https://example.com/x</loc>`)
	assert.False(t, strings.Contains(twice, `//https`))
}
