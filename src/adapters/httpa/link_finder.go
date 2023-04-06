package httpa

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"net/http"
	"strings"
)

func NewLinkFinder(httpClient *http.Client) *LinkFinder {
	return &LinkFinder{httpClient: httpClient}
}

type LinkFinder struct {
	httpClient *http.Client
}

func (l *LinkFinder) FindLinksOnPage(ctx context.Context, url domain.Link) ([]domain.Link, error) {
	doc, err := l.fetchDocument(ctx, url)
	if err != nil {
		return nil, err
	}

	return l.parseLinks(doc, url)
}

func (l *LinkFinder) parseLinks(doc *goquery.Document, pageURL domain.Link) (_ []domain.Link, returnErr error) {
	var links []domain.Link
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, hasHref := s.Attr("href")
		if !hasHref {
			return true
		}

		link, err := parseLink(pageURL, href)
		if err != nil {
			returnErr = err
			return false
		}

		links = append(links, link)
		return true
	})
	return links, returnErr
}

func (l *LinkFinder) fetchDocument(ctx context.Context, url domain.Link) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := l.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return doc, nil
}

func parseLink(pageURL domain.Link, href string) (domain.Link, error) {
	if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") {
		return domain.NewRelativeLink(pageURL, href)
	}

	return domain.NewLink(href)
}
