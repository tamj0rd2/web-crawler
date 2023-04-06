package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

func NewService(linkFinder LinkFinder, workerCount int) *Service {
	return &Service{
		linkFinder:  linkFinder,
		workerCount: workerCount,
	}
}

type Service struct {
	linkFinder  LinkFinder
	workerCount int
}

// LinkFinder will find the links on the given webpage
type LinkFinder interface {
	FindLinksOnPage(ctx context.Context, url Link) ([]Link, error)
}

func (s *Service) Crawl(ctx context.Context, startingURL Link, results chan<- VisitResult) error {
	var (
		wg           = &sync.WaitGroup{}
		links        = make(chan Link)
		visitedLinks = &sync.Map{}
	)

	for i := 0; i < s.workerCount; i++ {
		go s.visitLinks(ctx, wg, links, results, visitedLinks)
	}

	wg.Add(1)
	links <- startingURL
	wg.Wait()

	close(links)
	close(results)
	return nil
}

func (s *Service) visitLinks(ctx context.Context, activeJobs *sync.WaitGroup, linksToProcess chan Link, visits chan<- VisitResult, visitedURLs *sync.Map) {
	for pageURL := range linksToProcess {
		pageURL := pageURL.ToVisit()

		if _, alreadyVisited := visitedURLs.LoadOrStore(pageURL, true); alreadyVisited {
			activeJobs.Done()
			continue
		}

		linksOnPage, err := s.linkFinder.FindLinksOnPage(ctx, pageURL)
		if err != nil {
			visits <- VisitResult{Err: fmt.Errorf("failed to find links on %s: %w", pageURL, err)}
			activeJobs.Done()
			continue
		}

		visits <- VisitResult{Visit: NewVisit(pageURL, linksOnPage)}

		// add each visitable link to the queue
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
