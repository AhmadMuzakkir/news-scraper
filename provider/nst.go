package provider

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ahmadmuzakkir/scrapenews/model"
)

type Nst struct {
	httpClient *http.Client
	baseUrl    string
}

func NewNst(hc *http.Client) *Nst {
	return &Nst{
		httpClient: hc,
		baseUrl:    "https://www.nst.com.my",
	}
}

// TODO limit scraping by date (month) instead of maxPageNo ?
func (b *Nst) Scrape(source model.NewsSource, maxPageNo int, lastUpdate time.Time) ([]*model.News, error) {
	var list []*model.News
	var pageNo = 1

	for {
		page, err := b.scrapePage(source.Url+"?page="+strconv.Itoa(pageNo), source, lastUpdate)

		if page != nil {
			list = append(list, page...)
		}

		if err != nil && (err == ErrPageEmpty || err == ErrOldContent || (maxPageNo != 0 && pageNo == maxPageNo)) {
			break
		}

		if err != nil {
			return nil, err
		}

		pageNo++
	}

	return list, nil
}

func (b *Nst) scrapePage(url string, source model.NewsSource, lastUpdate time.Time) ([]*model.News, error) {
	doc, err := getUrl(b.httpClient, url)
	if err != nil {
		return nil, err
	}

	hasContent := doc.Find("div.view-content").Length() == 4
	if !hasContent {
		return nil, ErrPageEmpty
	}

	var newsList []*model.News
	var news *model.News
	doc.Find("div.views-row-inner").EachWithBreak(func(i int, s *goquery.Selection) bool {
		news = &model.News{Source: source}
		// Datetime

		//s.Find("div.views-field-created").Find("span").Each(func(i int, s *goquery.Selection) {
		//
		//	news.Datetime = s.Text()
		//	datetime, err2 := b.parseDate(s.Text())
		//	if err2 != nil {
		//		err = err2
		//		return
		//	}
		//
		//	news.Datetime = *datetime
		//
		//	log.Println("Datetime: ", news.Datetime)
		//})

		//if err != nil {
		//	return
		//}
		// Title
		s.Find("div.views-field-title").Find("a").Each(func(i int, s *goquery.Selection) {
			log.Println("Title: ", s.Text())
			news.Title = s.Text()

			val, exist := s.Attr("href")
			if !exist {
				err = ErrNoContent
				return
			}

			detailUrl := b.baseUrl + val
			log.Println("detail url: ", detailUrl)
			err = b.scrapeDetail(detailUrl, news)
			if err != nil {
				log.Println("scrapeDetail error: ", err)
				return
			}

			if news.Datetime.Before(lastUpdate) {
				err = ErrOldContent
				return
			}
		})

		if err != nil && (err == ErrNoContent || err == ErrOldContent) {
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

func (b *Nst) scrapeDetail(url string, news *model.News) error {
	var err error
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil
	}

	author := doc.Find("div.author").Find("a").Text()
	if author == "" {
		author = doc.Find("span.author").Find("a").Text()
	}
	author = strings.TrimSpace(author)
	log.Println("Author: ", author)
	news.Author = author

	datetime, err := b.parseDate(doc.Find("span.post-date").Text())
	if err != nil {
		return err
	}
	news.Datetime = *datetime

	var content string
	doc.Find("div.field-item.even").Find("p").Each(func(i int, s *goquery.Selection) {
		content += s.Text()
		content += "\n"
	})
	log.Println("Content: ", content)
	news.Content = content

	if locationIndex := strings.Index(content, ":"); locationIndex != -1 {
		news.Location = content[:locationIndex]
		news.Content = strings.TrimPrefix(content[locationIndex+1:], " ")

	} else {
		news.Content = content
	}

	var pictures []*model.Picture

	doc.Find("div.view.view-article-gallery").Each(func(i int, s *goquery.Selection) {
		imageUrl, exists := s.Find("div.views-field.views-field-field-image").Find("img").Attr("data-src")
		if !exists {
			return
		}
		log.Println("ImageUrl: ", imageUrl)

		caption := s.Find("div.views-field.views-field-field-image-caption").Find("div.field-content").Text()
		log.Println("Caption: ", caption)

		pictures = append(pictures, &model.Picture{ImageUrl: imageUrl, Caption: caption})
	})
	news.Pictures = pictures
	news.Url = url
	return nil
}

func (b *Nst) parseDate(str string) (*time.Time, error) {
	parts := strings.Fields(str)
	if len(parts) != 5 {
		return nil, errors.New(fmt.Sprintf("could not parse the date %s", str))
	}

	var month time.Month

	switch parts[0] {
	case "January":
		month = time.January
	case "February":
		month = time.February
	case "March":
		month = time.March
	case "April":
		month = time.April
	case "May":
		month = time.May
	case "June":
		month = time.June
	case "July":
		month = time.July
	case "August":
		month = time.August
	case "September":
		month = time.September
	case "October":
		month = time.October
	case "November":
		month = time.November
	case "December":
		month = time.December
	}

	if month == 0 {
		return nil, errors.New(fmt.Sprintf("could not parse the month %s", parts[0]))
	}

	var day int
	var err error

	dayPart := strings.Trim(parts[1], ",")
	if day, err = strconv.Atoi(dayPart); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the day %s", dayPart))
	}

	var year int

	yearPart := strings.TrimSpace(parts[2])
	if year, err = strconv.Atoi(yearPart); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the year %s", yearPart))
	}

	timePart := strings.TrimSpace(parts[4])
	clock := timePart[len(timePart)-2:]
	timeStr := timePart[:len(timePart)-2]

	timeArray := strings.Split(timeStr, ":")
	if len(timeArray) != 2 {
		return nil, errors.New(fmt.Sprintf("could not parse the time %s", timeStr))
	}

	var hour int
	if hour, err = strconv.Atoi(timeArray[0]); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the hour %s", timeArray[0]))
	}

	var minute int
	if minute, err = strconv.Atoi(timeArray[1]); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the minute %s", timeArray[1]))
	}

	if clock == "PM" || clock == "pm" {
		hour += 12
	}

	t := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)

	return &t, nil
}
