package model

import (
	"crypto/sha1"
	"fmt"
	"time"
)

type News struct {
	Id       string     `json:"id"`
	Author   string     `json:"author"`
	Datetime time.Time  `json:"datetime"`
	Title    string     `json:"title"`
	Location string     `json:"location"`
	Content  string     `json:"content"`
	Pictures []*Picture `json:"pictures"`
	Tags     []string   `json:"tags"`
	Url      string     `json:"url"`
	Source   NewsSource `json:"source"`
}

func (n *News) ToString() string {

	var tagStr string
	for _, t := range n.Tags {
		tagStr += t
		tagStr += ","
	}

	var pictureStr string
	for _, p := range n.Pictures {
		pictureStr += p.ToString()
		pictureStr += ","
	}

	return fmt.Sprintf(
		"{ author = %s, datetime = %v, title = %s, location = %s, tag = %s, picture = %s, content = %s",
		n.Author, n.Datetime, n.Title, n.Location, tagStr, pictureStr, n.Content,
	)
}

func (n *News) GenerateId() {
	hash := sha1.New()

	hash.Write([]byte(n.Url))

	n.Id = fmt.Sprintf("%x", hash.Sum(nil))
}

type Picture struct {
	ImageUrl string `json:"url"`
	Caption  string `json:"caption"`
}

func (p *Picture) ToString() string {
	return fmt.Sprintf(
		"{ imageurl = %s, Caption = %s",
		p.ImageUrl, p.Caption,
	)
}

func hashString(str string) string {
	hash := sha1.New()

	hash.Write([]byte(str))

	return fmt.Sprintf("%x", hash.Sum(nil))
}
