package httpa

import (
	"net/http"
	"time"
)

func NewHTTPClient(rateLimit time.Duration, timeout time.Duration) *http.Client {
	rateLimiter := time.Tick(rateLimit)
	httpClient := &http.Client{Timeout: timeout, Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		res, err := http.DefaultTransport.RoundTrip(req)
		<-rateLimiter
		return res, err
	})}
	return httpClient
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (r roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request)
}
