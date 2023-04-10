package domain_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"testing"
)

func TestNewLink(t *testing.T) {
	t.Run("trims the link and does not return an error if the link has spaces at either end", func(t *testing.T) {
		expectedOutputURL := "https://example.com"

		link, err := domain.NewAbsoluteLink("https://example.com ")
		require.NoError(t, err, "link with spaces at end should be parsable")
		assert.Equal(t, expectedOutputURL, link.String())

		link, err = domain.NewAbsoluteLink(" https://example.com")
		require.NoError(t, err, "link with spaces at start should be parsable")
		assert.Equal(t, expectedOutputURL, link.String())

		link, err = domain.NewAbsoluteLink(" https://example.com ")
		require.NoError(t, err, "link with spaces at both ends should be parsable")
		assert.Equal(t, expectedOutputURL, link.String())
	})
}
