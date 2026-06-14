package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
)

func TestDomainTableName(t *testing.T) {
	domain := &models.Domain{}
	if domain.TableName() != "domains" {
		t.Errorf("expected table name 'domains', got '%s'", domain.TableName())
	}
}

func TestAliasTableName(t *testing.T) {
	alias := &models.Alias{}
	if alias.TableName() != "aliases" {
		t.Errorf("expected table name 'aliases', got '%s'", alias.TableName())
	}
}

func TestReplyTokenTableName(t *testing.T) {
	token := &models.ReplyToken{}
	if token.TableName() != "reply_tokens" {
		t.Errorf("expected table name 'reply_tokens', got '%s'", token.TableName())
	}
}

func TestForwardLogTableName(t *testing.T) {
	log := &models.ForwardLog{}
	if log.TableName() != "forward_logs" {
		t.Errorf("expected table name 'forward_logs', got '%s'", log.TableName())
	}
}

func TestDomainStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	domain := &models.Domain{
		ID:        id,
		Name:      "test.com",
		Verified:  true,
		CreatedAt: now,
	}

	if domain.ID != id {
		t.Errorf("expected ID %v, got %v", id, domain.ID)
	}
	if domain.Name != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", domain.Name)
	}
	if domain.Verified != true {
		t.Errorf("expected verified true, got %v", domain.Verified)
	}
}

func TestAliasStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	alias := &models.Alias{
		ID:           id,
		Address:      "test@test.com",
		Slug:         "test-slug",
		Domain:       "test.com",
		RealEmail:    "real@example.com",
		DisplayName:  stringPtr("test-display-name"),
		Label:        stringPtr("test-label"),
		Enabled:      true,
		ForwardCount: 5,
		CreatedAt:    now,
		LastUsedAt:   &now,
	}

	if alias.ID != id {
		t.Errorf("expected ID %v, got %v", id, alias.ID)
	}
	if alias.Address != "test@test.com" {
		t.Errorf("expected address 'test@test.com', got '%s'", alias.Address)
	}
	if alias.Slug != "test-slug" {
		t.Errorf("expected slug 'test-slug', got '%s'", alias.Slug)
	}
	if alias.Domain != "test.com" {
		t.Errorf("expected domain 'test.com', got '%s'", alias.Domain)
	}
	if alias.RealEmail != "real@example.com" {
		t.Errorf("expected real_email 'real@example.com', got '%s'", alias.RealEmail)
	}
	if alias.DisplayName == nil || *alias.DisplayName != "test-display-name" {
		t.Errorf("expected display_name 'test-display-name', got %v", alias.DisplayName)
	}
	if alias.Label == nil || *alias.Label != "test-label" {
		t.Errorf("expected label 'test-label', got %v", alias.Label)
	}
	if alias.Enabled != true {
		t.Errorf("expected enabled true, got %v", alias.Enabled)
	}
	if alias.ForwardCount != 5 {
		t.Errorf("expected forward_count 5, got %d", alias.ForwardCount)
	}
}

func TestReplyTokenStruct(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(time.Hour)

	token := &models.ReplyToken{
		Token:           "test-token",
		AliasID:         uuid.New(),
		OriginalSender:  "sender@example.com",
		OriginalSubject: stringPtr("Test Subject"),
		ThreadID:        stringPtr("thread-123"),
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
	}

	if token.Token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", token.Token)
	}
	if token.OriginalSender != "sender@example.com" {
		t.Errorf("expected original_sender 'sender@example.com', got '%s'", token.OriginalSender)
	}
	if token.OriginalSubject == nil || *token.OriginalSubject != "Test Subject" {
		t.Errorf("expected original_subject 'Test Subject', got %v", token.OriginalSubject)
	}
	if token.ThreadID == nil || *token.ThreadID != "thread-123" {
		t.Errorf("expected thread_id 'thread-123', got %v", token.ThreadID)
	}
}

func TestForwardLogStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	log := &models.ForwardLog{
		ID:        id,
		AliasID:   uuid.New(),
		Direction: "inbound",
		Sender:    stringPtr("sender@example.com"),
		Subject:   stringPtr("Test Subject"),
		Status:    "delivered",
		CreatedAt: now,
	}

	if log.ID != id {
		t.Errorf("expected ID %v, got %v", id, log.ID)
	}
	if log.Direction != "inbound" {
		t.Errorf("expected direction 'inbound', got '%s'", log.Direction)
	}
	if log.Sender == nil || *log.Sender != "sender@example.com" {
		t.Errorf("expected sender 'sender@example.com', got %v", log.Sender)
	}
	if log.Subject == nil || *log.Subject != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got %v", log.Subject)
	}
	if log.Status != "delivered" {
		t.Errorf("expected status 'delivered', got '%s'", log.Status)
	}
}

func stringPtr(s string) *string {
	return &s
}
