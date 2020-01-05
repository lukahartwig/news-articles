package scraper

import (
	"fmt"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"github.com/sirupsen/logrus"

	"github.com/lukahartwig/news-articles/store"
)

type Scraper struct {
	Extractor Extractor
	Storage   *redisstorage.Storage
}

func (s *Scraper) Scrape(urls ...string) []store.Article {
	articles := make([]store.Article, 0)

	coll := colly.NewCollector(
		colly.AllowedDomains("www.spiegel.de", "www.bild.de", "deutsch.rt.com"),
	)

	err := coll.SetStorage(s.Storage)
	if err != nil {
		return nil
	}

	q, _ := queue.New(2, s.Storage)

	for _, url := range urls {
		q.AddURL(url)
	}

	coll.OnRequest(func(r *colly.Request) {
		logrus.Infof("visiting %s", r.URL)
	})

	coll.OnXML("//item", func(item *colly.XMLElement) {
		q.AddURL(item.ChildText("link"))
	})

	coll.OnHTML("html", func(elem *colly.HTMLElement) {
		articles = append(articles, store.Article{
			URL:          elem.Request.URL.String(),
			Publisher:    s.Extractor.Publisher(),
			Topic:        s.Extractor.Topic(elem),
			Headline:     s.Extractor.Headline(elem),
			Description:  s.Extractor.Description(elem),
			Keywords:     s.Extractor.Keywords(elem),
			Content:      s.Extractor.Content(elem),
			PublishedAt:  s.Extractor.PublishedAt(elem),
			LastModified: s.Extractor.LastModified(elem),
			Paywall:      s.Extractor.Paywall(elem),
		})
	})

	q.Run(coll)

	return articles
}

func meta(elem *colly.HTMLElement, name string) string {
	return elem.ChildAttr(fmt.Sprintf(`meta[name="%s"]`, name), "content")
}
