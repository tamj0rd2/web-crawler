package domain_test

import (
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"testing"
)

func TestCrawl(t *testing.T) {
	app := domain.NewService()
	spec.TestCrawl(t, app.Crawl)
}
