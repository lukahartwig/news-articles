package scraper

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"jaytaylor.com/html2text"
)

type SpiegelExtractor struct{}

func (ex SpiegelExtractor) Publisher() string {
	return "spiegel"
}

func (ex SpiegelExtractor) Topic(elem *colly.HTMLElement) string {
	return strings.Split(elem.Request.URL.Path, "/")[1]
}

func (ex SpiegelExtractor) Keywords(elem *colly.HTMLElement) []string {
	rawKeywords := strings.Split(meta(elem, "news_keywords"), ",")
	if len(rawKeywords) == 0 {
		rawKeywords = strings.Split(meta(elem, "keywords"), ",")
	}

	keywords := make([]string, 0, len(rawKeywords))
	for _, keyword := range rawKeywords {
		if k := strings.Trim(keyword, " "); k != "" {
			keywords = append(keywords, k)
		}
	}

	return keywords
}

func (ex SpiegelExtractor) Headline(elem *colly.HTMLElement) string {
	headline := elem.ChildText(".headline")
	if headline == "" {
		headline = elem.ChildText("h2")
	}
	return headline
}

func (ex SpiegelExtractor) Description(elem *colly.HTMLElement) string {
	return meta(elem, "description")
}

func (ex SpiegelExtractor) Content(elem *colly.HTMLElement) string {
	content, err := html2text.FromString(elem.ChildText(".article-section > p"), html2text.Options{
		OmitLinks: true,
	})
	if err != nil {
		logrus.Warnf("could not parse content: %s", err)
		return ""
	}
	return content
}

func (ex SpiegelExtractor) Paywall(elem *colly.HTMLElement) bool {
	return false
}

// iso8601 is a pattern to parse the ISO-8601 datetime format
const iso8601 = "2006-01-02T15:04:05-0700"

func (ex SpiegelExtractor) PublishedAt(elem *colly.HTMLElement) int64 {
	dateTime, err := time.Parse(iso8601, meta(elem, "date"))
	if err != nil {
		return 0
	}
	return dateTime.Unix()
}

func (ex SpiegelExtractor) LastModified(elem *colly.HTMLElement) int64 {
	dateTime, err := time.Parse(iso8601, meta(elem, "last-modified"))
	if err != nil {
		return 0
	}
	return dateTime.Unix()
}
