package main_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"github.com/tamj0rd2/web-crawler/src/spec"
	"os"
	"os/exec"
	"path"
	"testing"
)

func TestAcceptance(t *testing.T) {
	binary := path.Join(os.TempDir(), "crawler")
	require.NoError(t, exec.Command("go", "build", "-o", binary, "main.go").Run())
	defer os.Remove(binary)

	spec.TestCrawl(t, func(ctx context.Context, url domain.Link, visits chan<- domain.Visit) error {
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
			var visit domain.Visit
			if err := json.Unmarshal(scanner.Bytes(), &visit); err != nil {
				return fmt.Errorf("failed to unmarshal output: %w\noutput: %s", err, scanner.Text())
			}

			visits <- visit
		}

		if err := cmd.Wait(); err != nil {
			return err
		}

		close(visits)
		return nil
	})
}
