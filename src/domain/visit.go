package domain

import (
	"fmt"
	"net/url"
	"strings"
)

func NewVisit(pageURL Link, links []Link) Visit {
	return Visit{
		PageURL: pageURL,
		Links:   links,
	}
}

type VisitResult struct {
	Visit Visit
	Err   error
}

type Visit struct {
	PageURL Link
	Links   []Link
}

func NewLink(inputURL string) (Link, error) {
	parsedLink, err := url.Parse(strings.TrimSpace(inputURL))
	if err != nil {
		return "", fmt.Errorf("failed to parse link %s - %w", inputURL, err)
	}
	return Link(parsedLink.String()), nil
}

func NewRelativeLink(parent Link, path string) (Link, error) {
	parsedLink, err := parent.URL().Parse(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative link %s: %w", path, err)
	}
	return Link(parsedLink.String()), nil
}

// Link is a valid URL
type Link string

func (l Link) String() string {
	return string(l)
}

func (l Link) DomainName() string {
	return l.URL().Hostname()
}

func (l Link) ToVisit() Link {
	parsed := l.URL()
	parsed.Fragment = ""
	return Link(strings.TrimRight(parsed.String(), "/"))
}

func (l Link) URL() *url.URL {
	parsed, _ := url.Parse(l.String())
	return parsed
}
