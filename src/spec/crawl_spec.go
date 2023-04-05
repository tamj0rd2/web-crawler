package spec

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/domain/interactions"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCrawl(t *testing.T, crawl interactions.Crawl) {
	t.Run("Given a starting URL, each link is printed and visited recursively", func(t *testing.T) {
		const (
			homePath    = "/home"
			aboutPath   = "/about"
			contactPath = "/contact"
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == homePath {
				_, err := w.Write([]byte(`<html><body><a href="/about">About</a></body></html>`))
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			if r.URL.Path == aboutPath {
				_, err := w.Write([]byte(`<html><body><a href="/contact">Contact</a></body></html>`))
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			if r.URL.Path == contactPath {
				_, err := w.Write([]byte(`<html><body><a href="/home">Home</a></body></html>`))
				if err != nil {
					t.Fatal(err)
				}
				return
			}

			http.NotFound(w, r)
		}))
		defer server.Close()

		startingURL := server.URL + "/home"

		visits, err := crawl(context.Background(), domain.Link(startingURL))
		require.NoError(t, err)
		vh := visitsHelper{visits, server.URL}
		vh.assertLen(t, 3)
		vh.assertContains(t, homePath, []string{aboutPath})
		vh.assertContains(t, aboutPath, []string{contactPath})
		vh.assertContains(t, contactPath, []string{homePath})
	})
}

type visitsHelper struct {
	visits  []domain.Visit
	baseURL string
}

func (h visitsHelper) assertLen(t testing.TB, expected int) {
	t.Helper()
	assert.Len(t, h.visits, expected)
}

func (h visitsHelper) assertContains(t testing.TB, pageURL string, links []string) {
	t.Helper()

	if strings.HasPrefix(pageURL, "/") {
		pageURL = h.baseURL + pageURL
	}

	expectedLinks := make([]domain.Link, len(links))
	for i, link := range links {
		if strings.HasPrefix(link, "/") {
			expectedLinks[i] = domain.Link(h.baseURL + link)
		}
	}

	visitedLinks := make([]domain.Link, len(h.visits))
	for i, visit := range h.visits {
		if visit.Page == domain.Link(pageURL) {
			assert.ElementsMatch(t, visit.Links, expectedLinks, "expected listA, got listB")
			return
		}

		visitedLinks[i] = visit.Page
	}

	t.Errorf("%s was not visited. visited links: %v", pageURL, visitedLinks)
}
