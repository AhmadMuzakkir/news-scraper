package store

import (
	"log"
	"net/http"
	"time"

	"github.com/ahmadmuzakkir/scrapenews/model"
	"github.com/ahmadmuzakkir/scrapenews/provider"
	"github.com/pkg/errors"
)

var ErrProviderNotFound = errors.New("Provider not found !")

type Refresher struct {
	providers  map[string]provider.Provider
	store      NewsStore
	lastupdate time.Time
}

func NewRefresher(httpClient *http.Client, store NewsStore) *Refresher {
	refresh := &Refresher{}
	refresh.providers = make(map[string]provider.Provider)
	refresh.providers[model.NstId] = provider.NewNst(httpClient)
	refresh.providers[model.BharianId] = provider.NewBharian(httpClient)
	refresh.providers[model.UtusanId] = provider.NewUtusan(httpClient)

	refresh.lastupdate = time.Now().AddDate(0, 0, -1)
	refresh.store = store
	return refresh
}

func (r *Refresher) Refresh() {
	var workersCount = 10

	type result struct {
		news []*model.News
		err  error
	}

	jobs := make(chan model.NewsSource)
	results := make(chan result)

	// Starts the worker pools
	for w := 0; w < workersCount; w++ {
		go func() {
			for j := range jobs {
				p := r.providers[j.NewspaperId]

				if p == nil {
					results <- result{err: ErrProviderNotFound}
					continue
				}

				news, err := p.Scrape(j, 10, r.lastupdate)
				results <- result{news: news, err: err}
			}
		}()
	}

	var sources []model.NewsSource
	sources = append(sources, model.BhSources...)
	sources = append(sources, model.NstSources...)
	sources = append(sources, model.UtusanSources...)

	// Send the jobs to the workers
	go func() {
		for _, val := range sources {
			jobs <- val
		}

		close(jobs)
	}()

	// Get the results
	for i := 0; i < len(sources); i++ {
		res := <-results
		if res.err != nil {
			log.Println(res.err)
			continue
		}
		err := r.store.Insert(res.news)
		if err != nil {
			log.Println(err)
		}
	}
}
