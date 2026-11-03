package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is the subset of *http.Client this package needs, so tests can
// inject a fake instead of hitting the real Cloudflare API.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CloudflareClient talks to exactly one DNS record — this tool only ever
// has one job, so there's no zone/record lookup-by-name machinery here.
type CloudflareClient struct {
	HTTPClient HTTPClient
	Token      string
	ZoneID     string
	RecordID   string
}

type dnsRecord struct {
	Content string `json:"content"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type recordResponse struct {
	Result  dnsRecord  `json:"result"`
	Success bool       `json:"success"`
	Errors  []apiError `json:"errors"`
}

func (c *CloudflareClient) recordURL() string {
	return fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", c.ZoneID, c.RecordID)
}

func (c *CloudflareClient) do(method string, body io.Reader) (recordResponse, error) {
	req, err := http.NewRequest(method, c.recordURL(), body)
	if err != nil {
		return recordResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return recordResponse{}, err
	}
	defer res.Body.Close()

	var parsed recordResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return recordResponse{}, fmt.Errorf("decoding Cloudflare response: %w", err)
	}

	if !parsed.Success || res.StatusCode != http.StatusOK {
		return parsed, fmt.Errorf("cloudflare API error (HTTP %d): %v", res.StatusCode, parsed.Errors)
	}

	return parsed, nil
}

// CurrentContent returns the record's current `content` field (the IP it's
// presently pointed at).
func (c *CloudflareClient) CurrentContent() (string, error) {
	parsed, err := c.do(http.MethodGet, nil)
	if err != nil {
		return "", err
	}
	return parsed.Result.Content, nil
}

// UpdateContent PATCHes just the `content` field — a partial update, so the
// record's name/type/ttl/proxied settings are left untouched.
func (c *CloudflareClient) UpdateContent(newIP string) error {
	body := fmt.Sprintf(`{"content":%q}`, newIP)
	_, err := c.do(http.MethodPatch, strings.NewReader(body))
	return err
}
