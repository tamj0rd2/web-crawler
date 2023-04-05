package domain

import (
	"context"
	"fmt"
	"log"
)

func NewService(linkFinder LinkFinder) *Service {
	return &Service{
		seenURLs:   make(map[Link]bool),
		linkFinder: linkFinder,
	}
}

type Service struct {
	seenURLs   map[Link]bool
	linkFinder LinkFinder
}

type LinkFinder interface {
	FindLinksOnPage(ctx context.Context, url Link) ([]Link, error)
}

func (a Service) Crawl(ctx context.Context, startingURL Link) ([]Visit, error) {
	if a.seenURLs[startingURL] {
		return nil, nil
	}

	a.seenURLs[startingURL] = true

	links, err := a.linkFinder.FindLinksOnPage(ctx, startingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to find links on page %s - %w", startingURL, err)
	}

	log.Println("Visited", startingURL)
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
