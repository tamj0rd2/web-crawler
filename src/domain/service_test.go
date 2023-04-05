package domain_test

import (
	"github.com/tamj0rd2/web-crawler/src/adapters/httpa"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"net/http"
	"testing"
)

func TestCrawl(t *testing.T) {
	// improvement for the future: this should be able to be replaced with anything that implements the interface.
	// the test spec should be refactored to support this. For example, maybe we want to find links via cache or some
	// mechanism other than http
	linkFinder := httpa.NewLinkFinder(http.DefaultClient)

	app := domain.NewService(linkFinder, nil)
	spec.TestCrawl(t, app.Crawl)
}
