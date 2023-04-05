package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tamj0rd2/web-crawler/src/adapters/httpa"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"log"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	const (
		httpRateLimit = time.Second / 4
		httpTimeout   = time.Second * 5
	)

	httpClient := httpa.NewHTTPClient(httpRateLimit, httpTimeout)
	linkFinder := httpa.NewLinkFinder(httpClient)
	visitRecorder := domain.VisitRecorderFunc(func(ctx context.Context, visit domain.Visit) error {
		b, err := json.Marshal(visit)
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	})

	app := domain.NewService(linkFinder, visitRecorder)

	startingURL, err := domain.NewLink(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if _, err := app.Crawl(ctx, startingURL); err != nil {
		log.Fatal(err)
	}
}
