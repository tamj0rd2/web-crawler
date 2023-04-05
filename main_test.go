package main_test

import (
	"bytes"
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
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return nil, err
		}

		var visits []domain.Visit
		if err := json.Unmarshal(stdout.Bytes(), &visits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output: %w\noutput: %s", err, stdout.String())
		}

		return visits, nil
	})
}
