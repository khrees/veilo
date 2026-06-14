package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/resend/resend-go/v3"
)

func TestResendEmailProvider_RegisterDomain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer mock-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/domains" {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["name"] != "example.com" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resp := resend.Domain{
				Id:     "dom_123",
				Name:   "example.com",
				Status: "pending",
				Records: []resend.Record{
					{
						Type:  "MX",
						Name:  "mail",
						Value: "feedback-smtp.us-east-1.amazonses.com",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := resend.NewClient("mock-key")
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	prov := NewResendEmailProvider(client)
	res, err := prov.RegisterDomain(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.DomainID != "dom_123" {
		t.Errorf("expected dom_123, got %s", res.DomainID)
	}
	if len(res.Records) != 1 || res.Records[0].Type != "MX" || res.Records[0].Value != "feedback-smtp.us-east-1.amazonses.com" {
		t.Errorf("records translation mismatch: %+v", res.Records)
	}
}

func TestResendEmailProvider_VerifyDomain(t *testing.T) {
	var verifyCalled, getCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/domains/dom_123/verify" {
			verifyCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "GET" && r.URL.Path == "/domains/dom_123" {
			getCalled = true
			resp := resend.Domain{
				Id:     "dom_123",
				Status: "verified",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := resend.NewClient("mock-key")
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	prov := NewResendEmailProvider(client)
	verified, err := prov.VerifyDomain(context.Background(), "dom_123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !verified {
		t.Fatal("expected verified to be true")
	}
	if !verifyCalled || !getCalled {
		t.Errorf("expected verify and get endpoints to be called, got verify=%t, get=%t", verifyCalled, getCalled)
	}
}

func TestResendEmailProvider_GetReceivedEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: the SDK might call GET /emails/receiving/{id} or /emails/{id}
		// Let's verify what path is requested.
		resp := resend.Email{
			Id:      "email_123",
			Subject: "Hello",
			Html:    "<html>",
			Text:    "Hello",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resend.NewClient("mock-key")
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	prov := NewResendEmailProvider(client)
	email, err := prov.GetReceivedEmail(context.Background(), "email_123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if email.ID != "email_123" || email.Subject != "Hello" {
		t.Errorf("mismatch: %+v", email)
	}
}

func TestResendEmailProvider_SendEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/emails" {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["subject"] != "Hello" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resp := resend.SendEmailResponse{
				Id: "send_123",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := resend.NewClient("mock-key")
	u, _ := url.Parse(server.URL)
	client.BaseURL = u

	prov := NewResendEmailProvider(client)
	id, err := prov.SendEmail(context.Background(), SendEmailInput{
		From:    "sender@example.com",
		To:      []string{"receiver@example.com"},
		Subject: "Hello",
		Html:    "<html>",
		Text:    "Hello",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "send_123" {
		t.Errorf("expected send_123, got %s", id)
	}
}
