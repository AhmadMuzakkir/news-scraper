package provider

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ahmadmuzakkir/scrapenews/model"
	"github.com/pkg/errors"
)

type Bharian struct {
	httpClient *http.Client
	baseUrl    string
}

func NewBharian(hc *http.Client) *Bharian {
	return &Bharian{
		httpClient: hc,
		baseUrl:    "https://www.bharian.com.my",
	}
}

func (b *Bharian) Scrape(source model.NewsSource, maxPageNo int, lastUpdate time.Time) ([]*model.News, error) {
	var list []*model.News
	var pageNo = 1

	for {
		page, err := b.scrapePage(source.Url+"?page="+strconv.Itoa(pageNo), source, lastUpdate)

		if page != nil {
			list = append(list, page...)
		}

		if err != nil && (err == ErrPageEmpty || err == ErrOldContent || (maxPageNo > 0 && pageNo == maxPageNo)) {
			break
		}

		if err != nil {
			return nil, err
		}

		pageNo++
	}

	return list, nil
}

func (b *Bharian) scrapePage(url string, source model.NewsSource, lastUpdate time.Time) ([]*model.News, error) {
	doc, err := getUrl(b.httpClient, url)
	if err != nil {
		return nil, err
	}

	hasContent := doc.Find("div.view-content").Length() > 1
	if !hasContent {
		return nil, ErrPageEmpty
	}

	var newsList []*model.News
	var news *model.News
	doc.Find("div.views-row-inner").EachWithBreak(func(i int, s *goquery.Selection) bool {
		news = &model.News{Source: source}

		// Title
		s.Find("div.views-field.views-field-title").Find("span").Find("a").Each(func(i int, s *goquery.Selection) {
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

func (b *Bharian) scrapeDetail(url string, news *model.News) error {
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

	datetimeLabel := doc.Find("div.node-meta").Text()
	//datetimeLabel := strings.TrimSpace(doc.Find("div.node-meta").After("div").Text())
	datetime, err := b.parseDate(datetimeLabel)
	if err != nil {
		return err
	}
	news.Datetime = *datetime
	log.Println("datetime: ", news.Datetime)

	author := doc.Find("div.author").Text()
	log.Println("Author: ", author)
	author = strings.TrimPrefix(author, "Oleh")
	author = strings.TrimSpace(author)

	news.Author = author

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

func (b *Bharian) parseDate(str string) (*time.Time, error) {
	parts := strings.Fields(str)
	if len(parts) != 7 {
		return nil, errors.New(fmt.Sprintf("could not parse the date %d, %s", len(parts), str))
	}

	var month time.Month

	switch parts[3] {
	case "Januari":
		fallthrough
	case "JANUARI":
		fallthrough
	case "JAN":
		month = time.January
	case "Februari":
		fallthrough
	case "FEBRUARI":
		fallthrough
	case "FEB":
		month = time.February
	case "Mac":
		fallthrough
	case "MAC":
		month = time.March
	case "April":
		fallthrough
	case "APRIL":
		fallthrough
	case "APR":
		month = time.April
	case "Mei":
		fallthrough
	case "MEI":
		month = time.May
	case "Jun":
		fallthrough
	case "JUN":
		month = time.June
	case "Julai":
		fallthrough
	case "JULAI":
		fallthrough
	case "JUL":
		month = time.July
	case "Ogos":
		fallthrough
	case "OGOS":
		fallthrough
	case "OGO":
		month = time.August
	case "September":
		fallthrough
	case "SEPTEMBER":
		fallthrough
	case "SEP":
		month = time.September
	case "Oktober":
		fallthrough
	case "OKTOBER":
		fallthrough
	case "OKT":
		month = time.October
	case "November":
		fallthrough
	case "NOVEMBER":
		fallthrough
	case "NOV":
		month = time.November
	case "Disember":
		fallthrough
	case "DISEMBER":
		fallthrough
	case "DIS":
		month = time.December
	}

	if month == 0 {
		return nil, errors.New(fmt.Sprintf("could not parse the month %s", parts[2]))
	}

	var day int
	var err error

	dayPart := strings.Trim(parts[2], ",")
	if day, err = strconv.Atoi(dayPart); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the day %s", dayPart))
	}

	var year int

	yearPart := strings.TrimSpace(parts[4])
	if year, err = strconv.Atoi(yearPart); err != nil {
		return nil, errors.New(fmt.Sprintf("could not parse the year %s", yearPart))
	}

	timePart := strings.TrimSpace(parts[6])
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
