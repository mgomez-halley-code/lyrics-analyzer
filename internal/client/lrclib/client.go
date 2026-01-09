package lrclib

import (
	"fmt"
	"net/http"
	"time"
)

// Client handles communication with LRCLib API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new LRCLib client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// checkResponseStatus validates HTTP response status and returns appropriate errors
// This is shared across all endpoints
func (c *Client) checkResponseStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrLyricsNotFound

	case http.StatusInternalServerError:
		return &APIError{StatusCode: resp.StatusCode, Message: "LRCLib server error"}

	case http.StatusServiceUnavailable:
		return &APIError{StatusCode: resp.StatusCode, Message: "LRCLib service unavailable"}

	default:
		// Other 4xx errors (e.g., 400, 401, 403)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return &APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("Client error: %d", resp.StatusCode)}
		}
		// Other 5xx errors
		return &APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("Server error: %d", resp.StatusCode)}
	}
}
