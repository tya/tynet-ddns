package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// CurrentPublicIP asks a plain-text "what's my IP" endpoint — ipify's body
// *is* the IP, nothing to parse. Good enough for a home network behind one
// NAT; no need for anything smarter here.
func CurrentPublicIP(client HTTPClient) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ipify returned HTTP %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("ipify returned something that isn't an IP: %q", ip)
	}

	return ip, nil
}
