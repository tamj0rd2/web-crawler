package domain

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func NewService() *Service {
	return &Service{
		seenURLs: make(map[string]bool),
	}
}

type Service struct {
	seenURLs map[string]bool
}

func (a Service) Crawl(ctx context.Context, startingURL Link) ([]Visit, error) {
	if a.seenURLs[startingURL.String()] {
		return nil, nil
	}

	a.seenURLs[startingURL.String()] = true

	links, err := a.findUniqueLinksOnPage(ctx, startingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to find links on page %s - %w", startingURL, err)
	}

	visits := []Visit{{Page: startingURL, Links: links}}

	for _, link := range links {
		if link.DomainName() != startingURL.DomainName() {
			continue
		}

		linkVisits, err := a.Crawl(ctx, link)
		if err != nil {
			return nil, err
		}

		visits = append(visits, linkVisits...)
	}

	return visits, nil
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

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

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
			link, err := newRelativeLink(req.URL, href)
			if err != nil {
				// TODO: come back and handle this error
				panic(fmt.Errorf("failed to parse link %s: %w", href, err))
			}

			links = append(links, link)
			return
		}

		link, err := newLink(href)
		if err != nil {
			// TODO: come back and handle this error
			panic(fmt.Errorf("failed to parse link %s: %w", href, err))
		}

		links = append(links, link)
	})

	return links, nil
}
