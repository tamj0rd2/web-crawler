package domain

import (
	"context"
	"fmt"
	"log"
)

func NewService(linkFinder LinkFinder) *Service {
	return &Service{
		seenURLs:   make(map[string]bool),
		linkFinder: linkFinder,
	}
}

type Service struct {
	seenURLs   map[string]bool
	linkFinder LinkFinder
}

type LinkFinder interface {
	FindUniqueLinksOnPage(ctx context.Context, url Link) ([]Link, error)
}

func (a Service) Crawl(ctx context.Context, startingURL Link) ([]Visit, error) {
	if a.seenURLs[startingURL.String()] {
		return nil, nil
	}

	a.seenURLs[startingURL.String()] = true

	links, err := a.linkFinder.FindUniqueLinksOnPage(ctx, startingURL)
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
