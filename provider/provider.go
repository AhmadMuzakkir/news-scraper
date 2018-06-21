package provider

import (
	"github.com/ahmadmuzakkir/scrapenews/model"
	"time"
)

type Provider interface {
	Scrape(source model.NewsSource, maxPageNo int, lastUpdate time.Time) ([]*model.News, error)
}