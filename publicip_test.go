package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCurrentPublicIP(t *testing.T) {
	fake := &fakeHTTPClient{t: t, response: func(req *http.Request) *http.Response {
		if req.URL.String() != "https://api.ipify.org" {
			t.Fatalf("unexpected URL: %s", req.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("203.0.113.42\n"))}
	}}

	got, err := CurrentPublicIP(fake)
	if err != nil {
		t.Fatalf("CurrentPublicIP() error = %v", err)
	}
	if got != "203.0.113.42" {
		t.Errorf("CurrentPublicIP() = %q, want %q", got, "203.0.113.42")
	}
}

func TestCurrentPublicIP_NotAnIP(t *testing.T) {
	fake := &fakeHTTPClient{t: t, response: func(req *http.Request) *http.Response {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("<html>rate limited</html>"))}
	}}

	if _, err := CurrentPublicIP(fake); err == nil {
		t.Fatal("expected an error, got nil")
	}
}
