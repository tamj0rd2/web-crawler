package interactions

import (
	"context"
	"github.com/tamj0rd2/web-crawler/src/domain"
)

// Crawl will visit the starting URL and recursively visit all links found on the page
type Crawl func(ctx context.Context, url domain.Link, visits chan<- domain.VisitResult) error
