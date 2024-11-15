package entities

import (
	"time"

	amqp "github.com/kaellybot/kaelly-amqp"
)

type WebhookTwitter struct {
	WebhookID    string
	WebhookToken string
	TwitterID    string `gorm:"primaryKey"`
	Locale       amqp.Language
	RetryNumber  int64 `gorm:"default:0"`
	UpdatedAt    time.Time
}
