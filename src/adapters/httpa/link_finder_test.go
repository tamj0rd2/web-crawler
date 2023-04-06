package httpa_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tamj0rd2/web-crawler/src/adapters/httpa"
	"github.com/tamj0rd2/web-crawler/src/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLinkFinder(t *testing.T) {
	// NOTE: this behaviour could be improved if I had more time. I don't think there's a good reason to stop parsing
	// all links just because one of them is invalid
	t.Run("when a link on a page is invalid, an error is returned", func(t *testing.T) {
		const invalidLink = "http://  example.com/some-link"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(fmt.Sprintf(`<a href="%s">link</a>`, invalidLink)))
			require.NoError(t, err)
		}))
		defer server.Close()

		linkFinder := httpa.NewLinkFinder(server.Client())
		_, err := linkFinder.FindLinksOnPage(context.Background(), domain.Link(server.URL))
		require.Error(t, err)
	})
}
