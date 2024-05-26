package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// UnauthorizedError represents the message of an HTTP 401 response.
type UnauthorizedError ErrorMessage

// UnprocessableEntityError represents the generic structure of an error response.
type UnprocessableEntityError struct {
	Error ErrorMessage `json:"error"`
}

// ErrorMessage is the message of an error response.
type ErrorMessage struct {
	Message string `json:"message"`
}

// Client for the Hetzner DNS API.
type Client struct {
	requestLock sync.Mutex
	apiToken    string
	httpClient  *http.Client
	endPoint    *url.URL
}

// New creates a new API Client using a given api token.
func New(apiEndpoint string, apiToken string, maxRetires uint, httpClient *http.Client) (*Client, error) {
	endPoint, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing API endpoint URL: %w", err)
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := &Client{
		apiToken:   apiToken,
		endPoint:   endPoint,
		httpClient: httpClient,
	}

	client.httpClient.Transport = &retryableTransport{
		transport:  httpClient.Transport,
		maxRetries: int(maxRetires),
	}

	return client, nil
}

func (c *Client) request(ctx context.Context, method string, path string, bodyJSON any) (*http.Response, error) {
	uri := c.endPoint.JoinPath(path).String()

	tflog.Debug(ctx, fmt.Sprintf("HTTP request to API %s %s", method, uri))

	var (
		err     error
		reqBody []byte
	)

	if bodyJSON != nil {
		reqBody, err = json.Marshal(bodyJSON)
		if err != nil {
			return nil, fmt.Errorf("error serializing JSON body %s", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	// This lock ensures that only one request is sent to Hetzner API at a time.
	// See issue #5 for context.
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
		c.requestLock.Lock()
		defer c.requestLock.Unlock()
	}

	req.Header.Set("Auth-API-Token", c.apiToken)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		unauthorizedError, err := parseUnauthorizedError(resp)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("API returned HTTP 401 Unauthorized error with message: '%s'. Double check your API key is still valid", unauthorizedError.Message)
	} else if resp.StatusCode == http.StatusUnprocessableEntity {
		unprocessableEntityError, err := parseUnprocessableEntityError(resp)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("API returned HTTP 422 Unprocessable Entity error with message: '%s'", unprocessableEntityError.Error.Message)
	}

	return resp, nil
}
