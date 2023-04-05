package spec

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/domain/interactions"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrawl(t *testing.T, crawl interactions.Crawl) {
	t.Run("Given a starting URL, the links on the page are printed", func(t *testing.T) {
		const (
			homePath    = "/home"
			aboutPath   = "/about"
			contactPath = "/contact"
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == homePath {
				_, err := w.Write([]byte(`<html><body>
					<a href="/home">Home</a>
					<a href="/about">About</a>
					<a href="/contact">Contact</a>
				</body></html>`))
				if err != nil {
					t.Fatal(err)
				}
			}
		}))
		defer server.Close()

		startingURL := server.URL + "/home"

		links, err := crawl(context.Background(), domain.Link(startingURL))
		require.NoError(t, err)
		assert.Len(t, links, 3)
		assert.Contains(t, links, domain.Link(server.URL+homePath))
		assert.Contains(t, links, domain.Link(server.URL+aboutPath))
		assert.Contains(t, links, domain.Link(server.URL+contactPath))
	})
}
