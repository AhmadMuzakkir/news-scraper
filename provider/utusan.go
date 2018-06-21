package provider

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ahmadmuzakkir/scrapenews/model"
)

type Utusan struct {
	httpClient *http.Client
	baseUrl    string
	urls       []string
}

func NewUtusan(hc *http.Client) *Utusan {
	return &Utusan{
		httpClient: hc,
		baseUrl:    "http://www.utusan.com.my/",
		urls:       []string{"http://www.utusan.com.my/berita/nasional"},
	}
}

func (b *Utusan) Scrape(source model.NewsSource, maxPageNo int, lastUpdate time.Time) ([]*model.News, error) {
	var list []*model.News
	list, err := b.scrapePage(source, lastUpdate)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (b *Utusan) scrapePage(source model.NewsSource, lastUpdate time.Time) ([]*model.News, error) {
	log.Println("Scraping...")

	doc, err := getUrl(b.httpClient, source.Url)
	if err != nil {
		return nil, err
	}

	var newsList []*model.News
	var news *model.News
	doc.Find("li.element_item.item_teaser").Find("h2").Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		val, exist := s.Attr("href")
		if !exist {
			err = ErrNoContent
			return true
		}

		detailUrl := b.baseUrl + val
		log.Println("detail url: ", detailUrl)

		news = &model.News{Source: source}
		err := b.scrapeDetail(detailUrl, news)
		if err != nil {
			return true
		}

		if news.Datetime.Before(lastUpdate) {
			return false
		}

		news.GenerateId()
		newsList = append(newsList, news)
		log.Println()
		return true
	})

	if err != nil {
		return nil, err
	}

	return newsList, nil
}

func (b *Utusan) scrapeDetail(url string, news *model.News) error {
	doc, err := getUrl(b.httpClient, url)
	if err != nil {
		return err
	}

	title := doc.Find("div.content_header.content__header.tonal__header").Find("h1").Text()
	news.Title = title
	log.Println("title: ", title)

	var pictures []*model.Picture

	doc.Find("#lightbox-links").Find("a").Each(func(i int, s *goquery.Selection) {
		var url string
		var caption string
		var exist bool

		if url, exist = s.Attr("href"); !exist {
			return
		}
		log.Println("url: ", url)

		url = b.baseUrl + url
		caption, _ = s.Attr("title")

		log.Println("caption: ", caption)

		pictures = append(pictures, &model.Picture{ImageUrl: url, Caption: caption})

	})
	news.Pictures = pictures

	if timestamp, exist := doc.Find("p.content__dateline").Find("time").Attr("data-timestamp"); exist {
		log.Println("timestamp: ", timestamp)

		timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			log.Println("timestamp err: ", err)
			return err
		}
		news.Datetime = time.Unix(timestampInt/1000, 0)
	}

	news.Author = doc.Find("a.tone-colour.author").Find("span").Text()

	var content string
	doc.Find("div.clearfix.article_body.content__article-body.from-content-api.js-article__body").Find("p").Each(func(i int, s *goquery.Selection) {
		content += s.Text()
		content += "\n"
	})
	log.Println("content: ", content)
	news.Content = content

	var tags []string
	doc.Find("ul.tag-list").Find("li").Each(func(i int, s *goquery.Selection) {
		tag := s.Find("a").Text()

		log.Println("tag: ", tag)
		tags = append(tags, tag)
	})
	news.Tags = tags
	news.Url = url
	log.Println("news: ", news.ToString())
	log.Println()

	return nil
}
