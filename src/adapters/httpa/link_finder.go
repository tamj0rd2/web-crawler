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

func (l *LinkFinder) FindUniqueLinksOnPage(ctx context.Context, url domain.Link) ([]domain.Link, error) {
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
		return nil, err
	}

	var links []domain.Link
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, hasHref := s.Attr("href")
		if !hasHref {
			return
		}

		if strings.HasPrefix(href, "/") {
			link, err := domain.NewRelativeLink(req.URL, href)
			if err != nil {
				// TODO: come back and handle this error
				panic(fmt.Errorf("failed to parse relative link %s: %w", href, err))
			}

			links = append(links, link)
			return
		}

		link, err := domain.NewLink(href)
		if err != nil {
			// TODO: come back and handle this error
			panic(fmt.Errorf("failed to parse link %s: %w", href, err))
		}

		links = append(links, link)
	})

	return links, nil
}
