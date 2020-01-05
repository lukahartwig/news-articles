package scraper

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"jaytaylor.com/html2text"
)

type RTExtractor struct{}

func (ex RTExtractor) Publisher() string {
	return "russiatoday"
}

func (ex RTExtractor) Topic(elem *colly.HTMLElement) string {
	return strings.Split(elem.Request.URL.Path, "/")[1]
}

func (ex RTExtractor) Keywords(elem *colly.HTMLElement) []string {
	keywords := make([]string, 0)
	elem.ForEach(".tags__heading ~ a", func(_ int, tag *colly.HTMLElement) {
		if k := strings.Trim(tag.Text, " \n\t\r"); k != "" {
			keywords = append(keywords, k)
		}
	})
	return keywords
}

func (ex RTExtractor) Headline(elem *colly.HTMLElement) string {
	return elem.ChildText(".article__heading")
}

func (ex RTExtractor) Description(elem *colly.HTMLElement) string {
	return meta(elem, "description")
}

func (ex RTExtractor) Content(elem *colly.HTMLElement) string {
	content, err := html2text.FromString(elem.ChildText(".article__text > p"), html2text.Options{
		OmitLinks: true,
	})
	if err != nil {
		logrus.Warnf("could not parse article content: %s", err)
		return ""
	}
	return content
}

func (ex RTExtractor) Paywall(elem *colly.HTMLElement) bool {
	return false
}

const rtDateTime = "2.01.2006 • 15:04 Uhr"

func (ex RTExtractor) PublishedAt(elem *colly.HTMLElement) int64 {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return 0
	}

	dateTime, err := time.ParseInLocation(rtDateTime, elem.ChildText(".article__date"), loc)
	if err != nil {
		return 0
	}
	return dateTime.Unix()
}

func (ex RTExtractor) LastModified(elem *colly.HTMLElement) int64 {
	return 0
}
