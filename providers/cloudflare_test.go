package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudflareDNSProvider_ConfigureDNS_Success(t *testing.T) {
	// Mock Cloudflare API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mock-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method == "GET" && r.URL.Path == "/client/v4/zones" {
			if r.URL.Query().Get("name") != "example.com" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			resp := map[string]any{
				"result": []map[string]any{
					{"id": "zone-123"},
				},
				"success": true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/client/v4/zones/zone-123/dns_records" {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["name"] != "mail.example.com" || body["type"] != "MX" || body["content"] != "feedback-smtp.us-east-1.amazonses.com" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{"success": true})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	prov := &cloudflareDNSProvider{
		token:   "mock-token",
		baseURL: server.URL,
	}

	records := []DNSRecord{
		{
			Type:     "MX",
			Name:     "mail.example.com",
			Value:    "feedback-smtp.us-east-1.amazonses.com",
			Priority: 10,
		},
	}

	err := prov.ConfigureDNS(context.Background(), "example.com", records)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCloudflareDNSProvider_ConfigureDNS_NoZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"result":  []map[string]any{},
			"success": true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	prov := &cloudflareDNSProvider{
		token:   "mock-token",
		baseURL: server.URL,
	}

	err := prov.ConfigureDNS(context.Background(), "example.com", []DNSRecord{})
	if err == nil {
		t.Fatal("expected error since no zone was found, got nil")
	}
}

func TestCloudflareDNSProvider_ConfigureDNS_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	prov := &cloudflareDNSProvider{
		token:   "mock-token",
		baseURL: server.URL,
	}

	err := prov.ConfigureDNS(context.Background(), "example.com", []DNSRecord{})
	if err == nil {
		t.Fatal("expected error due to HTTP 500, got nil")
	}
}
