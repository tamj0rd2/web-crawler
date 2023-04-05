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

	urls, err := app.Crawl(ctx, domain.Link(os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(urls)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(string(b))
}
