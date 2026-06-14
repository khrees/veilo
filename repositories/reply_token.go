package repositories

import (
	"time"

	"github.com/khrees/veilo/models"
	"gorm.io/gorm"
)

type ReplyTokenRepository interface {
	Create(t *models.ReplyToken) error
	FindByToken(token string) (*models.ReplyToken, error)
	Delete(token string) error
	DeleteExpired(now time.Time) error
}

type replyTokenRepository struct {
	db *gorm.DB
}

func NewReplyTokenRepository(db *gorm.DB) ReplyTokenRepository {
	return &replyTokenRepository{db: db}
}

func (r *replyTokenRepository) Create(t *models.ReplyToken) error {
	return r.db.Create(t).Error
}

func (r *replyTokenRepository) FindByToken(token string) (*models.ReplyToken, error) {
	var rt models.ReplyToken
	return &rt, r.db.First(&rt, "token = ?", token).Error
}

func (r *replyTokenRepository) Delete(token string) error {
	return r.db.Delete(&models.ReplyToken{}, "token = ?", token).Error
}

func (r *replyTokenRepository) DeleteExpired(now time.Time) error {
	return r.db.Where("expires_at < ?", now).Delete(&models.ReplyToken{}).Error
}
