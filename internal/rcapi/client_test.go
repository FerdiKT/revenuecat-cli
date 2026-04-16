package rcapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientRetriesReadRequests(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Bearer sk_test"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		call := calls.Add(1)
		if call == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message":    "rate limited",
				"retryable":  true,
				"backoff_ms": 0,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "ok"})
	}))
	defer server.Close()

	client := NewClient("sk_test", server.URL)
	result, err := client.Do(context.Background(), Request{
		Method:    http.MethodGet,
		Path:      "projects",
		RetryMode: RetryDefault,
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	if got, want := calls.Load(), int32(2); got != want {
		t.Fatalf("calls = %d, want %d", got, want)
	}
	payload := result.Payload.(map[string]any)
	if payload["id"] != "ok" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestClientReturnsAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":    "authorization_error",
			"message": "forbidden",
		})
	}))
	defer server.Close()

	client := NewClient("sk_test", server.URL)
	_, err := client.Do(context.Background(), Request{
		Method:    http.MethodGet,
		Path:      "projects",
		RetryMode: RetryDefault,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Type != "authorization_error" {
		t.Fatalf("apiErr = %#v", apiErr)
	}
}
