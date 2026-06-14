package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type cloudflareDNSProvider struct {
	token   string
	baseURL string
}

// NewCloudflareDNSProvider creates a new DNSProvider backed by Cloudflare REST API
func NewCloudflareDNSProvider(token string) DNSProvider {
	return &cloudflareDNSProvider{
		token:   token,
		baseURL: "https://api.cloudflare.com",
	}
}

func (p *cloudflareDNSProvider) ConfigureDNS(ctx context.Context, domainName string, records []DNSRecord) error {
	if p.token == "" {
		return fmt.Errorf("cloudflare API token is not configured")
	}

	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://api.cloudflare.com"
	}

	// 1. Lookup Zone ID by domain name
	reqURL := fmt.Sprintf("%s/client/v4/zones?name=%s", baseURL, url.QueryEscape(domainName))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudflare zones lookup failed: status %d", resp.StatusCode)
	}

	var zoneResp struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&zoneResp); err != nil {
		return err
	}

	if len(zoneResp.Result) == 0 {
		return fmt.Errorf("no Cloudflare zone found for domain %s", domainName)
	}
	zoneID := zoneResp.Result[0].ID

	// 2. Create DNS records on Cloudflare
	for _, rec := range records {
		bodyData := map[string]any{
			"type":    rec.Type,
			"name":    rec.Name,
			"content": rec.Value,
			"ttl":     3600,
			"proxied": false,
		}
		if rec.Type == "MX" {
			bodyData["priority"] = rec.Priority
		}

		jsonBody, err := json.Marshal(bodyData)
		if err != nil {
			return err
		}

		postURL := fmt.Sprintf("%s/client/v4/zones/%s/dns_records", baseURL, zoneID)
		postReq, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return err
		}
		postReq.Header.Set("Authorization", "Bearer "+p.token)
		postReq.Header.Set("Content-Type", "application/json")

		postResp, err := client.Do(postReq)
		if err != nil {
			return err
		}
		postResp.Body.Close()

		if postResp.StatusCode < 200 || postResp.StatusCode >= 300 {
			return fmt.Errorf("cloudflare dns record creation failed: status %d", postResp.StatusCode)
		}
	}

	return nil
}
