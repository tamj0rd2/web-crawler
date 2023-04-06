# Web crawler

## How to run:

### Prerequisites:

- install go version 1.19
- run `go mod tidy` to install dependencies

### Crawling a website

Run `go run main.go <starting_url>` in a terminal to start crawling a page

Example: `go run ./main.go https://monzo.com`

The output of successful visits are printed to stdout whilst errors are printed to stderr.

### How to test:

Run `go test ./... --race` in a terminal to run all tests, or `make test`

**A note on testing**: I used a top-down approach to test the CLI and my service. As the tests run so quickly (especially without the race flag), it was quick and easy for me to test most behaviours at the top level as acceptance tests. Due to this, the test coverage in the lower layers (for example example, `link_finder.go`) seems thin. If this was a real project and features/functionality started growing and causing the tests to become expensive to run, I would prioritise testing the edge cases and failure scenarios further down the testing pyramid.

## Improvements

These things could be improved if I had more time:

- Make requests per second, rate limits and timeouts etc configurable via the CLI
- When parsing a link on a page fails, we give up processing that page. There's no reason for a page to be completely skipped just because parsing a link on the page failed.
- Respect robots.txt
- Handle pages that would have content rendered client-side, such as SPAs
- Not every file type needs to be visited. I've excluded pdfs and mp3s, but the exclusions could be expanded
- The constraint around which links to visit/not visit could be injected into the service as a function
- As the service layer expects a `LinkFinder` which is not coupled to implementation, `crawl_spec.go` should not need to be coupled with http

## Decisions along the way

My first pass of the functionality was to recursively visit pages and return all the visits as a slice at the end. Once I started doing exploratory testing (to discover edge cases + unknowns) against https://monzo.com, it very quickly became clear that keeping everything in memory and printing at the very end was a bad idea.

Crawling sites is slow, especially when attempting to be a good internet citizen. Waiting until the end of the crawl to print out visits meant that I had no visibility of what was going on, without additional logging. Also, depending on the amount of links that need to be crawled, holding all of those links in memory would eat up the computer's resources for no good reason.

So rather than holding everything in memory to print at the end, I decided to print each Visit to stdout immediately. Once I implemented concurrency, this further developed into sending the visits to a channel for receivers to consume however they want. This allowed me to decouple presentation logic (in this case, printing to stdout) from the service layer.
