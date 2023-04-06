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
	log.SetFlags(log.Ltime)

	const (
		requestsPerSecond = 4
		httpRateLimit     = time.Second / requestsPerSecond
		httpTimeout       = time.Second * 15
		workerCount       = requestsPerSecond * 2
	)

	httpClient := httpa.NewHTTPClient(httpRateLimit, httpTimeout)
	linkFinder := httpa.NewLinkFinder(httpClient)
	app := domain.NewService(linkFinder, workerCount)

	startingURL, err := domain.NewLink(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	results := make(chan domain.VisitResult)
	done := make(chan bool)
	go func() {
		for visit := range results {
			if visit.Err != nil {
				log.Println(visit.Err)
				continue
			}

			b, _ := json.Marshal(visit)
			fmt.Println(string(b))
		}
		done <- true
	}()

	if err := app.Crawl(context.Background(), startingURL, results); err != nil {
		log.Fatal(err)
	}

	<-done
	log.Println("Done!")
}
