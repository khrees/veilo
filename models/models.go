// Package models contains the data models for the application.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Domain struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"type:text;not null;uniqueIndex"`
	Verified  bool           `json:"verified" gorm:"not null;default:false"`
	CreatedAt time.Time      `json:"created_at" gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Domain) TableName() string { return "domains" }

func (d *Domain) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type Alias struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Address      string     `json:"address" gorm:"type:text;not null;uniqueIndex"`
	Slug         string     `json:"slug" gorm:"type:text;not null"`
	Domain       string     `json:"domain" gorm:"type:text;not null"`
	RealEmail    string     `json:"real_email" gorm:"type:text;not null"`
	DisplayName  *string    `json:"display_name,omitempty" gorm:"type:text"`
	Label        *string    `json:"label,omitempty" gorm:"type:text"`
	Enabled      bool       `json:"enabled" gorm:"not null"`
	ForwardCount int        `json:"forward_count" gorm:"not null;default:0"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null;default:now()"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	MaxForwards  *int       `json:"max_forwards,omitempty"`
}

type AliasFilter struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Domain  *string `json:"domain,omitempty"`
	Limit   *int    `json:"limit,omitempty"`
	Offset  *int    `json:"offset,omitempty"`
}

func (Alias) TableName() string { return "aliases" }

func (a *Alias) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type ReplyToken struct {
	Token           string    `json:"token" gorm:"type:text;primaryKey"`
	AliasID         uuid.UUID `json:"alias_id" gorm:"type:uuid;not null;index"`
	OriginalSender  string    `json:"original_sender" gorm:"type:text;not null"`
	OriginalSubject *string   `json:"original_subject,omitempty" gorm:"type:text"`
	ThreadID        *string   `json:"thread_id,omitempty" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;default:now()"`
	ExpiresAt       time.Time `json:"expires_at" gorm:"not null"`
}

func (ReplyToken) TableName() string { return "reply_tokens" }

type ForwardLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AliasID   uuid.UUID `json:"alias_id" gorm:"type:uuid;not null;index"`
	Direction string    `json:"direction" gorm:"type:text;not null;check:direction IN ('inbound','reply')"`
	Sender    *string   `json:"sender,omitempty" gorm:"type:text"`
	Subject   *string   `json:"subject,omitempty" gorm:"type:text"`
	Status          string    `json:"status" gorm:"type:text;not null;check:status IN ('delivered','blocked','bounced')"`
	TrackersBlocked int       `json:"trackers_blocked" gorm:"not null;default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;default:now();index"`
}

func (ForwardLog) TableName() string { return "forward_logs" }

func (f *ForwardLog) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

type Stats struct {
	TotalAliases         int64 `json:"total_aliases"`
	TotalForwarded       int64 `json:"total_forwarded"`
	TotalBlocked         int64 `json:"total_blocked"`
	TotalTrackersBlocked int64 `json:"total_trackers_blocked"`
}

