package main_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"os"
	"os/exec"
	"testing"
)

func TestAcceptance(t *testing.T) {
	spec.TestCrawl(t, func(ctx context.Context, url domain.Link) ([]domain.Visit, error) {
		cmd := exec.CommandContext(ctx, "go", "run", "main.go", string(url))
		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command: %w", err)
		}

		var visits []domain.Visit
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var visit domain.Visit
			if err := json.Unmarshal(scanner.Bytes(), &visit); err != nil {
				return nil, fmt.Errorf("failed to unmarshal output: %w\noutput: %s", err, scanner.Text())
			}
			visits = append(visits, visit)
		}

		if err := cmd.Wait(); err != nil {
			return nil, err
		}

		return visits, nil
	})
}
