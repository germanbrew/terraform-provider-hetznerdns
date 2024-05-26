package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func parseUnprocessableEntityError(resp *http.Response) (*UnprocessableEntityError, error) {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("error reading HTTP response body: %e", err)
	}

	var unprocessableEntityError UnprocessableEntityError

	err = json.Unmarshal(body, &unprocessableEntityError)
	if err != nil {
		return nil, err
	}

	return &unprocessableEntityError, nil
}

func parseUnauthorizedError(resp *http.Response) (*UnauthorizedError, error) {
	var unauthorizedError UnauthorizedError

	err := readAndParseJSONBody(resp, &unauthorizedError)
	if err != nil {
		return nil, err
	}

	return &unauthorizedError, nil
}

func readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return body, err
}

func readAndParseJSONBody(resp *http.Response, respType interface{}) error {
	body, err := readBody(resp)
	if err != nil {
		return fmt.Errorf("error reading HTTP response body %w", err)
	}

	return json.Unmarshal(body, &respType)
}
