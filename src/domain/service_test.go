package domain_test

import (
	"context"
	"fmt"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"strings"
	"sync"
	"testing"
)

func TestCrawl(t *testing.T) {
	linkFinder := &inMemLinkFinder{baseURL: "http://localhost:1234"}
	app := domain.NewService(linkFinder, 3)
	spec.TestCrawl(t, app.Crawl, linkFinder.setupFixture)
}

type inMemLinkFinder struct {
	baseURL     domain.Link
	fixtureLock sync.Mutex
	fixture     spec.Fixture
}

func (f *inMemLinkFinder) FindLinksOnPage(ctx context.Context, url domain.Link) ([]domain.Link, error) {
	path := strings.TrimPrefix(url.String(), f.baseURL.String())

	f.fixtureLock.Lock()
	hrefs, found := f.fixture[path]
	if !found {
		f.fixtureLock.Unlock()
		return nil, fmt.Errorf("no fixture found for path %s", path)
	}
	f.fixtureLock.Unlock()

	linksToReturn := make([]domain.Link, len(hrefs))
	for i, href := range hrefs {
		link, err := domain.NewLink(f.baseURL, href)
		if err != nil {
			return nil, err
		}
		linksToReturn[i] = link
	}

	return linksToReturn, nil
}

func (f *inMemLinkFinder) setupFixture(t testing.TB, fixture spec.Fixture) (baseURL string) {
	f.fixtureLock.Lock()
	defer f.fixtureLock.Unlock()
	f.fixture = fixture
	return f.baseURL.String()
}
