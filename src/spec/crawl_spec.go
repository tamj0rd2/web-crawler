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
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{aboutPath})
		vh.assertContains(t, aboutPath, []string{contactPath})
		vh.assertContains(t, contactPath, []string{homePath})
	})

	t.Run("Pages on the same domain are visited and printed", func(t *testing.T) {
		const (
			homePath         = "/home"
			pathOnSameDomain = "/path-on-same-domain"
		)

		serverRoutes := routes{}
		server := httptest.NewUnstartedServer(serverRoutes)

		urlOnSameDomain := server.URL + pathOnSameDomain
		serverRoutes[homePath] = htmlWithLinks(urlOnSameDomain)
		serverRoutes[pathOnSameDomain] = htmlWithLinks(homePath)

		server.Start()
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{urlOnSameDomain})
		vh.assertContains(t, pathOnSameDomain, []string{homePath})
	})

	t.Run("Pages on different domains are not visited, but they are printed", func(t *testing.T) {
		const (
			homePath             = "/home"
			urlOnDifferentDomain = "https://example.com/something"
		)

		serverRoutes := routes{
			homePath: htmlWithLinks(urlOnDifferentDomain),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "maybe the external path is being visited by mistake?")
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{urlOnDifferentDomain})
	})

	t.Run("Pages on different sub-domains are not visited, but they are printed", func(t *testing.T) {
		const (
			homePath                = "/home"
			urlOnDifferentSubDomain = "http://subdomain.localhost/something"
		)

		serverRoutes := routes{
			homePath: htmlWithLinks(urlOnDifferentSubDomain),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "maybe the external path is being visited by mistake?")
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{urlOnDifferentSubDomain})
	})

	t.Run("Links that appear with and without trailing slashes are only visited once", func(t *testing.T) {
		const (
			homePath = "/home"
		)

		serverRoutes := routes{
			homePath: htmlWithLinks(homePath + "/"),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "if the error is a 404, maybe the link with the trailing slash being visited by mistake?")
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{homePath})
	})

	t.Run("It can handle nav links and footers without visiting the links multiple times", func(t *testing.T) {
		const (
			homePath    = "/home"
			aboutPath   = "/about"
			contactPath = "/contact"
		)

		serverRoutes := routes{
			homePath:    htmlWithLinks(homePath, aboutPath, contactPath),
			aboutPath:   htmlWithLinks(homePath, aboutPath, contactPath),
			contactPath: htmlWithLinks(homePath, aboutPath, contactPath),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath
		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{homePath, aboutPath, contactPath})
		vh.assertContains(t, aboutPath, []string{homePath, aboutPath, contactPath})
		vh.assertContains(t, contactPath, []string{homePath, aboutPath, contactPath})
	})

	t.Run("It lists anchors and will only visit the linked page once", func(t *testing.T) {
		const (
			homePath  = "/home"
			aboutPath = "/about"
			// TODO: missing a test here for hrefs that are just anchors
			contactAnchor = "/about#contact"
		)

		serverRoutes := routes{
			homePath:  htmlWithLinks(homePath, aboutPath, contactAnchor),
			aboutPath: htmlWithLinks(homePath, aboutPath, contactAnchor),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath
		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{homePath, aboutPath, contactAnchor})
		vh.assertContains(t, aboutPath, []string{homePath, aboutPath, contactAnchor})
	})

	t.Run("pdf and mp3s are listed but not visited", func(t *testing.T) {
		const (
			homePath = "/home"
			mp3Path  = "/something.mp3"
			pdfPath  = "/something.pdf"
		)

		serverRoutes := routes{
			homePath: htmlWithLinks(homePath, mp3Path, pdfPath),
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.assertLen(t, len(serverRoutes))
		vh.assertContains(t, homePath, []string{homePath, mp3Path, pdfPath})
	})
}

func newVisitsHelper(baseURL string) (*visitsHelper, chan bool) {
	vh := &visitsHelper{baseURL: baseURL, results: make(chan domain.Visit)}

	done := make(chan bool)
	go func() {
		for visit := range vh.results {
			vh.visits = append(vh.visits, visit)
		}
		done <- true
	}()

	return vh, done
}

type visitsHelper struct {
	visits  []domain.Visit
	baseURL string
	results chan domain.Visit
}

func (h visitsHelper) assertLen(t testing.TB, expected int) {
	t.Helper()
	assert.Len(t, h.visits, expected)
}

func (h visitsHelper) assertContains(t testing.TB, pageURL string, links []string) {
	t.Helper()
	expectedPageLink := toLink(h.baseURL, pageURL)
	expectedLinks := toLinks(h.baseURL, links...)

	visitedLinks := make([]domain.Link, len(h.visits))
	for i, visit := range h.visits {
		if visit.PageURL == expectedPageLink {
			assert.ElementsMatch(t, visit.Links, expectedLinks, "expected listA, got listB")
			return
		}

		visitedLinks[i] = visit.PageURL
	}

	t.Errorf("%s was not visited. visited links: %v", expectedPageLink, visitedLinks)
}

func toLink(baseURL string, path string) domain.Link {
	if strings.HasPrefix(path, "/") {
		return domain.Link(baseURL + path)
	}
	return domain.Link(path)
}

func toLinks(baseURL string, paths ...string) []domain.Link {
	links := make([]domain.Link, len(paths))
	for i, path := range paths {
		links[i] = toLink(baseURL, path)
	}
	return links
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
