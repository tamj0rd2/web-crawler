package interactions

import (
	"context"
	"github.com/tamj0rd2/web-crawler/src/domain"
)

type Crawl func(ctx context.Context, url domain.Link) ([]domain.Visit, error)
