package domain

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func NewService() *Service {
	return &Service{}
}

type Service struct {
}

func (a Service) Crawl(ctx context.Context, startingURL Link) ([]Link, error) {
	links, err := a.findUniqueLinksOnPage(ctx, startingURL)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (a Service) findUniqueLinksOnPage(ctx context.Context, url Link) ([]Link, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var links []Link
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, hasHref := s.Attr("href")
		if !hasHref {
			return
		}

		if strings.HasPrefix(href, "/") {
			parsedLink, err := req.URL.Parse(href)
			if err != nil {
				// TODO: come back and handle this error
				panic(fmt.Errorf("failed to parse link %s: %w", href, err))
			}

			href = parsedLink.String()
		}

		links = append(links, Link(href))
	})

	return links, nil
}
