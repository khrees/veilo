package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/khrees/veilo/models"
	"github.com/khrees/veilo/repositories"
)

// IReplyTokenService interface for reply token operations
type IReplyTokenService interface {
	Create(input ReplyTokenCreateInput) (*models.ReplyToken, error)
	Find(token string) (*models.ReplyToken, error)
	Delete(token string) error
}

// ReplyTokenCreateInput groups the values needed to create a reply token.
type ReplyTokenCreateInput struct {
	AliasID         uuid.UUID
	OriginalSender  string
	OriginalSubject string
	ThreadID        string
	ExpiresAt       time.Time
}

// replyTokenService implements IReplyTokenService
type replyTokenService struct {
	replyTokenRepo repositories.ReplyTokenRepository
}

// NewReplyTokenService will instantiate ReplyTokenService
func NewReplyTokenService(replyTokenRepo repositories.ReplyTokenRepository) IReplyTokenService {
	return &replyTokenService{
		replyTokenRepo: replyTokenRepo,
	}
}

// Create creates a new reply token
func (r *replyTokenService) Create(input ReplyTokenCreateInput) (*models.ReplyToken, error) {
	token := uuid.NewString()

	var subject *string
	if input.OriginalSubject != "" {
		subject = &input.OriginalSubject
	}
	var threadID *string
	if input.ThreadID != "" {
		threadID = &input.ThreadID
	}

	replyToken := &models.ReplyToken{
		Token:           token,
		AliasID:         input.AliasID,
		OriginalSender:  input.OriginalSender,
		OriginalSubject: subject,
		ThreadID:        threadID,
		ExpiresAt:       input.ExpiresAt,
	}

	err := r.replyTokenRepo.Create(replyToken)
	if err != nil {
		return nil, err
	}

	return replyToken, nil
}

// Find finds a reply token
func (r *replyTokenService) Find(token string) (*models.ReplyToken, error) {
	return r.replyTokenRepo.FindByToken(token)
}

// Delete deletes a reply token
func (r *replyTokenService) Delete(token string) error {
	return r.replyTokenRepo.Delete(token)
}
