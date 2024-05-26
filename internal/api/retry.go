package api

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type retryableTransport struct {
	transport  http.RoundTripper
	maxRetries int
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.transport == nil {
		t.transport = http.DefaultTransport
	}

	// Clone the request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	// Send the request
	resp, err := t.transport.RoundTrip(req)
	// Retry logic
	retries := 0
	for shouldRetry(err, resp) && retries < t.maxRetries {
		// Wait for the specified backoff period
		time.Sleep(backoff(retries))
		// We're going to retry, consume any response to reuse the connection.
		drainBody(resp)
		// Clone the request body again

		tflog.Debug(req.Context(), fmt.Sprintf("Request to API %s %s", req.Method, req.URL))

		if req.Body != nil {
			tflog.Trace(req.Context(), string(bodyBytes))

			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Retry the request
		resp, err = t.transport.RoundTrip(req)

		if resp != nil {
			var b bytes.Buffer
			resp.Body = io.NopCloser(io.TeeReader(resp.Body, &b))

			tflog.Debug(req.Context(), fmt.Sprintf("HTTP response from API %s %s", resp.Status, req.URL))
			tflog.Trace(req.Context(), fmt.Sprintf("%+v", resp.Header))
			tflog.Trace(req.Context(), b.String())
		}

		retries++
	}

	return resp, err
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if rateLimit := resp.Header.Get("X-Ratelimit-Remaining-Minute"); rateLimit == "0" {
		time.Sleep(time.Minute)
	}

	return resp.StatusCode == http.StatusUnprocessableEntity
}

func drainBody(resp *http.Response) {
	if resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries/2))) * time.Second
}
