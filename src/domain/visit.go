package domain

import (
	"fmt"
	"net/url"
)

type Visit struct {
	Page  Link
	Links []Link
}

func newLink(inputURL string) (Link, error) {
	parsedLink, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse link %s: %w", inputURL, err)
	}
	return Link(parsedLink.String()), nil
}

func newRelativeLink(base *url.URL, path string) (Link, error) {
	parsedLink, err := base.Parse(path)
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
	parsed, _ := url.Parse(l.String())
	return parsed.Hostname()
}
