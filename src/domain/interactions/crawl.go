package interactions

import (
	"context"
	"github.com/tamj0rd2/web-crawler/src/domain"
)

type Crawl func(ctx context.Context, url domain.Link, visits chan<- domain.Visit) error
