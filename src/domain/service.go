package domain

import (
	"context"
	"fmt"
)

func NewService(linkFinder LinkFinder, visitRecorder VisitRecorder) *Service {
	if visitRecorder == nil {
		visitRecorder = noopVisitRecorder
	}

	return &Service{
		seenURLs:      make(map[Link]bool),
		linkFinder:    linkFinder,
		visitRecorder: visitRecorder,
	}
}

type Service struct {
	seenURLs      map[Link]bool
	linkFinder    LinkFinder
	visitRecorder VisitRecorder
}

type LinkFinder interface {
	FindLinksOnPage(ctx context.Context, url Link) ([]Link, error)
}

type VisitRecorder interface {
	RecordVisit(ctx context.Context, visit Visit) error
}

type VisitRecorderFunc func(ctx context.Context, visit Visit) error

func (v VisitRecorderFunc) RecordVisit(ctx context.Context, visit Visit) error {
	return v(ctx, visit)
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

	visit := Visit{Page: startingURL, Links: links}
	if err := a.visitRecorder.RecordVisit(ctx, visit); err != nil {
		return nil, fmt.Errorf("failed to record visit - %w", err)
	}

	visits := []Visit{visit}

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

var noopVisitRecorder = VisitRecorderFunc(func(ctx context.Context, visit Visit) error {
	return nil
})
