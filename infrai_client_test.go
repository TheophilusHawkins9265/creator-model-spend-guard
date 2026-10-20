package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRejectedEnvelopeIncludesStatusAndDetail(t *testing.T) {
	for _, tc := range []struct {
		name, body, detail string
	}{
		{"missing error", `{"ok":false}`, "no error detail"},
		{"null error", `{"ok":false,"error":null}`, "no error detail"},
		{"provided error", `{"ok":false,"error":{"code":"DENIED"}}`, "DENIED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &InfraiClient{HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}}
			err := client.call(context.Background(), http.MethodPut, "/account/budget/set", nil, nil)
			if err == nil || !strings.Contains(err.Error(), "HTTP 502") || !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}
