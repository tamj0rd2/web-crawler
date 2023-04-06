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
	app := domain.NewService(linkFinder)

	startingURL, err := domain.NewLink(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	visits := make(chan domain.Visit)
	done := make(chan bool)
	go func() {
		for visit := range visits {
			b, _ := json.Marshal(visit)
			fmt.Println(string(b))
		}
		done <- true
	}()

	if err := app.Crawl(ctx, startingURL, visits); err != nil {
		log.Fatal(err)
	}

	<-done
	log.Println("Done!")
}
