package domain_test

import (
	"github.com/tamj0rd2/web-crawler/src/adapters/httpa"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"net/http"
	"testing"
)

func TestCrawl(t *testing.T) {
	linkFinder := httpa.NewLinkFinder(http.DefaultClient)
	app := domain.NewService(linkFinder, 1)
	spec.TestCrawl(t, app.Crawl)
}
