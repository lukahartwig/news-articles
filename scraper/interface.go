package scraper

import (
	"github.com/gocolly/colly"
)

type Extractor interface {
	Publisher() string
	Topic(elem *colly.HTMLElement) string
	Keywords(elem *colly.HTMLElement) []string
	Headline(elem *colly.HTMLElement) string
	Description(elem *colly.HTMLElement) string
	Content(elem *colly.HTMLElement) string
	Paywall(elem *colly.HTMLElement) bool
	PublishedAt(elem *colly.HTMLElement) int64
	LastModified(elem *colly.HTMLElement) int64
}
