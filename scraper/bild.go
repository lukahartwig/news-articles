package scraper

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
	"jaytaylor.com/html2text"
)

type BildExtractor struct{}

func (ex BildExtractor) Publisher() string {
	return "bild"
}

func (ex BildExtractor) Topic(elem *colly.HTMLElement) string {
	seg := strings.Split(elem.Request.URL.Path, "/")
	if seg[1] == "bild-plus" {
		return seg[2]
	}
	return seg[1]
}

func (ex BildExtractor) Keywords(elem *colly.HTMLElement) []string {
	str := meta(elem, "news_keywords")
	if str == "" {
		str = meta(elem, "keywords")
	}
	rawKeywords := strings.Split(str, ",")

	keywords := make([]string, 0, len(rawKeywords))
	for _, keyword := range rawKeywords {
		if k := strings.Trim(keyword, " "); k != "" {
			keywords = append(keywords, k)
		}
	}

	return keywords
}

func (ex BildExtractor) Headline(elem *colly.HTMLElement) string {
	headline := elem.ChildText("#cover")
	if headline != "" {
		return headline
	}
	return elem.ChildText("h1:first-of-type")
}

func (ex BildExtractor) Description(elem *colly.HTMLElement) string {
	return meta(elem, "description")
}

func (ex BildExtractor) Content(elem *colly.HTMLElement) string {
	str := elem.ChildText(".txt > p")
	if str == "" {
		str = elem.ChildText(".article-body > p")
	}

	content, err := html2text.FromString(str, html2text.Options{
		OmitLinks: true,
	})
	if err != nil {
		logrus.Warnf("could not parse article content: %s", err)
		return ""
	}
	return content
}

func (ex BildExtractor) Paywall(elem *colly.HTMLElement) bool {
	return strings.Contains(elem.Request.URL.String(), "bild-plus")
}

func (ex BildExtractor) PublishedAt(elem *colly.HTMLElement) int64 {
	str := elem.ChildAttr(".authors__pubdate", "datetime")
	if str == "" {
		str = elem.ChildAttr(".authors > time", "datetime")
	}
	dateTime, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return 0
	}
	return dateTime.Unix()
}

func (ex BildExtractor) LastModified(elem *colly.HTMLElement) int64 {
	return 0
}
