package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Ltime)
	ctx := context.Background()
	app := domain.NewService()

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
