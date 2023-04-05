package spec

import (
	"context"
	"fmt"
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

		serverRoutes := routes{
			homePath:    htmlWithLinks(aboutPath),
			aboutPath:   htmlWithLinks(contactPath),
			contactPath: htmlWithLinks(homePath),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + "/home"

		visits, err := crawl(context.Background(), domain.Link(startingURL))
		require.NoError(t, err)

		vh := visitsHelper{visits, server.URL}
		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{aboutPath})
		vh.assertContains(t, aboutPath, []string{contactPath})
		vh.assertContains(t, contactPath, []string{homePath})
	})

	t.Run("Pages on different domains are not visited, but they are printed", func(t *testing.T) {
		const (
			homePath     = "/home"
			externalPath = "https://example.com"
		)

		serverRoutes := routes{
			homePath: htmlWithLinks(externalPath),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + "/home"

		visits, err := crawl(context.Background(), domain.Link(startingURL))
		require.NoError(t, err, "maybe the external path is being visited by mistake?")

		vh := visitsHelper{visits, server.URL}
		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{externalPath})
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
		expectedLinks[i] = domain.Link(link)

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

type routes map[string]string

func (routes routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html, found := routes[r.URL.Path]
	if !found {
		http.NotFound(w, r)
		return
	}

	if _, err := w.Write([]byte(html)); err != nil {
		panic(fmt.Errorf("could not write response: %w", err))
	}
}

func htmlWithLinks(links ...string) string {
	for i, link := range links {
		links[i] = fmt.Sprintf(`<a href="%s">%s</a>`, link, link)
	}
	return fmt.Sprintf(`<html><body>%s</body></html>`, strings.Join(links, "\n"))
}
