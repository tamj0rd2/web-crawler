package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

func NewService(linkFinder LinkFinder) *Service {
	return &Service{
		linkFinder: linkFinder,
	}
}

type Service struct {
	linkFinder LinkFinder
}

type LinkFinder interface {
	FindLinksOnPage(ctx context.Context, url Link) ([]Link, error)
}

func (s *Service) Crawl(ctx context.Context, startingURL Link, visits chan<- Visit) error {
	const workerCount = 3
	linksToProcess := make(chan Link)

	visitedLinks := &sync.Map{}
	activeJobs := &sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		go s.visitLinks(ctx, activeJobs, linksToProcess, visits, visitedLinks)
	}

	activeJobs.Add(1)
	linksToProcess <- startingURL
	activeJobs.Wait()

	close(linksToProcess)
	close(visits)
	return nil
}

func (s *Service) visitLinks(ctx context.Context, activeJobs *sync.WaitGroup, linksToProcess chan Link, visits chan<- Visit, visitedURLs *sync.Map) {
	for pageURL := range linksToProcess {
		pageURL := pageURL.WithoutAnchor()

		if _, alreadyVisited := visitedURLs.LoadOrStore(pageURL, true); alreadyVisited {
			activeJobs.Done()
			continue
		}

		linksOnPage, err := s.linkFinder.FindLinksOnPage(ctx, pageURL)
		if err != nil {
			// TODO: handle this error
			panic(fmt.Errorf("failed to find links on page %s: %w", pageURL, err))
		}

		visits <- NewVisit(pageURL, linksOnPage)

		go func() {
			for _, link := range linksOnPage {
				if !s.canVisit(pageURL, link) {
					continue
				}

				activeJobs.Add(1)
				linksToProcess <- link
			}
			activeJobs.Done()
		}()
	}
}

func (s *Service) canVisit(parentLink, link Link) bool {
	if link.DomainName() != parentLink.DomainName() {
		return false
	}

	if str := link.String(); strings.HasSuffix(str, ".pdf") || strings.HasSuffix(str, ".mp3") {
		return false
	}

	return true
}
