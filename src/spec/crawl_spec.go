package spec

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/domain/interactions"
	"log"
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
			homePath:    {aboutPath},
			aboutPath:   {contactPath},
			contactPath: {homePath},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("Pages on the same domain are visited and printed", func(t *testing.T) {
		const (
			homePath         = "/home"
			pathOnSameDomain = "/path-on-same-domain"
		)

		serverRoutes := routes{}
		server := httptest.NewUnstartedServer(serverRoutes)

		urlOnSameDomain := server.URL + pathOnSameDomain
		serverRoutes[homePath] = []string{urlOnSameDomain}
		serverRoutes[pathOnSameDomain] = []string{homePath}

		server.Start()
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("Pages on different domains are not visited, but they are printed", func(t *testing.T) {
		const (
			homePath             = "/home"
			urlOnDifferentDomain = "https://example.com/something"
		)

		serverRoutes := routes{
			homePath: {urlOnDifferentDomain},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "maybe the external path is being visited by mistake?")
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("Pages on different sub-domains are not visited, but they are printed", func(t *testing.T) {
		const (
			homePath                = "/home"
			urlOnDifferentSubDomain = "http://subdomain.localhost/something"
		)

		serverRoutes := routes{
			homePath: {urlOnDifferentSubDomain},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "maybe the external path is being visited by mistake?")
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("Links that appear with and without trailing slashes are only visited once", func(t *testing.T) {
		const (
			homePath                  = "/home"
			homePathWithTrailingSlash = homePath + "/"
		)

		serverRoutes := routes{
			homePath: {homePathWithTrailingSlash},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()
		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err, "if the error is a 404, maybe the link with the trailing slash being visited by mistake?")
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("It can handle nav links and footers without visiting the links multiple times", func(t *testing.T) {
		const (
			homePath    = "/home"
			aboutPath   = "/about"
			contactPath = "/contact"
		)

		serverRoutes := routes{
			homePath:    {homePath, aboutPath, contactPath},
			aboutPath:   {homePath, aboutPath, contactPath},
			contactPath: {homePath, aboutPath, contactPath},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath
		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	t.Run("It lists anchors and will only visit the linked page once", func(t *testing.T) {
		const (
			homePath               = "/home"
			aboutPath              = "/about"
			contactAnchor          = "/about#contact"
			contactAnchorWithSlash = "/about/#contact"
		)

		serverRoutes := routes{
			homePath:  {homePath, aboutPath, contactAnchor, contactAnchorWithSlash},
			aboutPath: {homePath},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath
		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.AssertMatches(t, serverRoutes)
	})

	// excluding these specific ones because I saw them a lot during testing, and they slow the program down
	t.Run("pdf and mp3s are listed but not visited", func(t *testing.T) {
		const (
			homePath = "/home"
			mp3Path  = "/something.mp3"
			pdfPath  = "/something.pdf"
		)

		serverRoutes := routes{
			homePath: {homePath, mp3Path, pdfPath},
		}

		server := httptest.NewServer(serverRoutes)
		defer server.Close()

		startingURL := server.URL + homePath

		vh, done := newVisitsHelper(server.URL)
		err := crawl(context.Background(), domain.Link(startingURL), vh.results)
		require.NoError(t, err)
		<-done

		vh.AssertMatches(t, serverRoutes)
	})
}

func newVisitsHelper(baseURL string) (*visitsHelper, chan bool) {
	vh := &visitsHelper{baseURL: baseURL, results: make(chan domain.VisitResult)}

	done := make(chan bool)
	go func() {
		for result := range vh.results {
			if result.Err != nil {
				panic(result.Err)
			}

			log.Println(result.Visit)
			vh.visits = append(vh.visits, result.Visit)
		}
		done <- true
	}()

	return vh, done
}

type visitsHelper struct {
	visits  []domain.Visit
	baseURL string
	results chan domain.VisitResult
}

func (h visitsHelper) assertLen(t testing.TB, expected int) {
	t.Helper()
	assert.Len(t, h.visits, expected)
}

func (h visitsHelper) assertContains(t testing.TB, pageURL string, links ...string) {
	t.Helper()
	expectedPageLink := toLink(h.baseURL, pageURL)
	expectedLinks := toLinks(h.baseURL, links...)

	visitedLinks := make([]domain.Link, len(h.visits))
	for i, visit := range h.visits {
		if visit.PageURL == expectedPageLink {
			assert.ElementsMatch(t, expectedLinks, visit.Links, "links for %s did not match expectations. expected listA, got listB", visit.PageURL)
			return
		}

		visitedLinks[i] = visit.PageURL
	}

	t.Errorf("%s was not visited. visited links: %v", expectedPageLink, visitedLinks)
}

func (h visitsHelper) AssertMatches(t *testing.T, serverRoutes routes) {
	t.Helper()
	h.assertLen(t, len(serverRoutes))
	expectedVisits := serverRoutes.asLinks(h.baseURL)

	for _, visit := range h.visits {
		routeLinks, found := expectedVisits[visit.PageURL]
		if !found {
			t.Errorf("unexpected page visited: %s", visit.PageURL)
			continue
		}
		assert.ElementsMatch(t, routeLinks, visit.Links, "expected listA, got listB")
	}
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

// routes is a map of server paths to links. The links can be either absolute or relative.
type routes map[string][]string

func (routes routes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	links, found := routes[r.URL.Path]
	if !found {
		http.NotFound(w, r)
		return
	}

	htmlLinks := make([]string, len(links))
	for i, link := range links {
		htmlLinks[i] = fmt.Sprintf(`<a href="%s">%s</a>`, link, link)
	}
	html := fmt.Sprintf(`<html><body>%s</body></html>`, strings.Join(htmlLinks, "\n"))

	if _, err := w.Write([]byte(html)); err != nil {
		panic(fmt.Errorf("could not write response: %w", err))
	}
}

func (routes routes) asLinks(baseURL string) map[domain.Link][]domain.Link {
	converted := make(map[domain.Link][]domain.Link)
	for page, links := range routes {
		converted[toLink(baseURL, page)] = toLinks(baseURL, links...)
	}
	return converted
}
