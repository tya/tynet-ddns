package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeHTTPClient struct {
	t        *testing.T
	response func(req *http.Request) *http.Response
	lastBody string
}

func (f *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			f.t.Fatalf("reading request body: %v", err)
		}
		f.lastBody = string(b)
	}
	return f.response(req), nil
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestCurrentContent(t *testing.T) {
	fake := &fakeHTTPClient{t: t, response: func(req *http.Request) *http.Response {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected Bearer test-token, got %q", got)
		}
		return jsonResponse(http.StatusOK, `{"success":true,"result":{"content":"1.2.3.4"}}`)
	}}

	cf := &CloudflareClient{HTTPClient: fake, Token: "test-token", ZoneID: "zone1", RecordID: "rec1"}

	got, err := cf.CurrentContent()
	if err != nil {
		t.Fatalf("CurrentContent() error = %v", err)
	}
	if got != "1.2.3.4" {
		t.Errorf("CurrentContent() = %q, want %q", got, "1.2.3.4")
	}
}

func TestCurrentContent_APIError(t *testing.T) {
	fake := &fakeHTTPClient{t: t, response: func(req *http.Request) *http.Response {
		return jsonResponse(http.StatusForbidden, `{"success":false,"errors":[{"code":9109,"message":"Invalid access token"}]}`)
	}}

	cf := &CloudflareClient{HTTPClient: fake, Token: "bad", ZoneID: "zone1", RecordID: "rec1"}

	if _, err := cf.CurrentContent(); err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestUpdateContent(t *testing.T) {
	fake := &fakeHTTPClient{t: t, response: func(req *http.Request) *http.Response {
		if req.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", req.Method)
		}
		return jsonResponse(http.StatusOK, `{"success":true,"result":{"content":"5.6.7.8"}}`)
	}}

	cf := &CloudflareClient{HTTPClient: fake, Token: "test-token", ZoneID: "zone1", RecordID: "rec1"}

	if err := cf.UpdateContent("5.6.7.8"); err != nil {
		t.Fatalf("UpdateContent() error = %v", err)
	}
	if !strings.Contains(fake.lastBody, `"content":"5.6.7.8"`) {
		t.Errorf("request body = %q, want it to contain the new IP", fake.lastBody)
	}
}
