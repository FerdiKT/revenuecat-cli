package rcapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type RetryMode int

const (
	RetryDefault RetryMode = iota
	RetryDisabled
	RetryForced
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

type Request struct {
	Method    string
	Path      string
	Query     url.Values
	Body      any
	RetryMode RetryMode
}

type Result struct {
	StatusCode int
	RequestID  string
	Payload    any
}

type APIError struct {
	Type       string `json:"type,omitempty"`
	Param      string `json:"param,omitempty"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable,omitempty"`
	DocURL     string `json:"doc_url,omitempty"`
	BackoffMS  *int   `json:"backoff_ms,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
	}
}

func (c *Client) Do(ctx context.Context, req Request) (*Result, error) {
	if c.baseURL == "" {
		return nil, errors.New("api base url is required")
	}
	if c.apiKey == "" {
		return nil, errors.New("api key is required")
	}

	fullURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	fullURL.Path = path.Join(fullURL.Path, req.Path)
	if len(req.Query) > 0 {
		fullURL.RawQuery = req.Query.Encode()
	}

	var body []byte
	if req.Body != nil {
		body, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}

	attempts := 1
	if shouldRetry(req.Method, req.RetryMode) {
		attempts = 4
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL.String(), bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
		httpReq.Header.Set("Accept", "application/json")
		if req.Body != nil {
			httpReq.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if attempt == attempts {
				return nil, err
			}
			sleepWithJitter(backoffDuration(attempt, nil))
			continue
		}

		result, apiErr := decodeResponse(resp)
		if apiErr == nil {
			return result, nil
		}
		if attempt == attempts || !retryableStatus(resp.StatusCode, apiErr) {
			return nil, apiErr
		}
		sleepWithJitter(backoffDuration(attempt, apiErr.BackoffMS))
	}

	return nil, errors.New("unexpected retry loop exit")
}

func shouldRetry(method string, mode RetryMode) bool {
	if mode == RetryForced {
		return true
	}
	if mode == RetryDisabled {
		return false
	}
	return method == http.MethodGet
}

func retryableStatus(status int, apiErr *APIError) bool {
	if apiErr != nil && apiErr.Retryable {
		return true
	}
	switch status {
	case http.StatusTooManyRequests, http.StatusLocked:
		return true
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func backoffDuration(attempt int, backoffMS *int) time.Duration {
	if backoffMS != nil && *backoffMS > 0 {
		return time.Duration(*backoffMS) * time.Millisecond
	}
	pow := math.Pow(2, float64(attempt-1))
	return time.Duration(pow*250) * time.Millisecond
}

func sleepWithJitter(d time.Duration) {
	jitter := time.Duration(rand.IntN(150)) * time.Millisecond
	time.Sleep(d + jitter)
}

func decodeResponse(resp *http.Response) (*Result, *APIError) {
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			Message:    fmt.Sprintf("read response body: %v", err),
			StatusCode: resp.StatusCode,
		}
	}

	requestID := resp.Header.Get("X-Request-Id")
	if requestID == "" {
		requestID = resp.Header.Get("X-Request-ID")
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if len(data) > 0 {
			_ = json.Unmarshal(data, &apiErr)
		}
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
		apiErr.StatusCode = resp.StatusCode
		return nil, &apiErr
	}

	if resp.StatusCode == http.StatusNoContent || len(data) == 0 {
		return &Result{StatusCode: resp.StatusCode, RequestID: requestID, Payload: map[string]any{}}, nil
	}

	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, &APIError{
			Message:    fmt.Sprintf("decode response body: %v", err),
			StatusCode: resp.StatusCode,
		}
	}

	return &Result{
		StatusCode: resp.StatusCode,
		RequestID:  requestID,
		Payload:    payload,
	}, nil
}
