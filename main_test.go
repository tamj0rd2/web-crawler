package main_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestAcceptance(t *testing.T) {
	binary := path.Join(os.TempDir(), "crawler")
	require.NoError(t, exec.Command("go", "build", "-o", binary, "main.go").Run())
	defer os.Remove(binary)

	spec.TestCrawl(t, newCrawler(binary), setupFixture)
}

func newCrawler(binary string) func(ctx context.Context, url domain.Link, results chan<- domain.VisitResult) error {
	return func(ctx context.Context, url domain.Link, results chan<- domain.VisitResult) error {
		cmd := exec.CommandContext(ctx, binary, string(url))
		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var result domain.VisitResult
			if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
				return fmt.Errorf("failed to unmarshal output: %w\noutput: %s", err, scanner.Text())
			}

			results <- result
		}

		if err := cmd.Wait(); err != nil {
			return err
		}

		close(results)
		return nil
	}
}

func setupFixture(t testing.TB, fixture spec.Fixture) (baseURL string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		links, found := fixture[r.URL.Path]
		if !found {
			http.NotFound(w, r)
			return
		}

		htmlLinks := make([]string, len(links))
		for i, link := range links {
			htmlLinks[i] = fmt.Sprintf(`<a href="%s">%s</a>`, link, link)
		}
		html := fmt.Sprintf(`<html><body>%s</body></html>`, strings.Join(htmlLinks, "\n"))

		if _, err := w.Write([]byte(html)); err != nil {
			panic(fmt.Errorf("could not write response: %w", err))
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}
