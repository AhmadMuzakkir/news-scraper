package store

import (
	"time"

	"github.com/ahmadmuzakkir/scrapenews/model"
)

type NewsStore interface {
	Insert(news []*model.News) error
	GetByKeywords(keywords []string) ([]*model.News, error)
	GetAll(from time.Time, end time.Time) ([]*model.News, error)
	GetByProvider(providerId string) ([]*model.News, error)
}
