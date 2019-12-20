package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
	"github.com/sirupsen/logrus"
	"jaytaylor.com/html2text"

	"github.com/lukahartwig/news-articles/store"
)

const feedPattern = "https://www.spiegel.de/%s/index.rss"

func scrape(config Config) ([]store.Article, error) {
	articles := make([]store.Article, 0)

	coll := colly.NewCollector()

	storage := &redisstorage.Storage{
		Address: config.RedisAddr,
		Prefix:  "spiegel",
		Expires: config.RedisExpire,
	}

	err := coll.SetStorage(storage)
	if err != nil {
		return nil, errors.New("failed to set storage")
	}

	defer storage.Client.Close()

	q, _ := queue.New(2, storage)

	for _, topic := range config.Topics {
		q.AddURL(fmt.Sprintf(feedPattern, topic))
	}

	coll.OnRequest(func(r *colly.Request) {
		logrus.Infof("visiting %s", r.URL)
	})

	coll.OnXML("//item", func(item *colly.XMLElement) {
		link := item.ChildText("link")
		q.AddURL(strings.Replace(link, ".html", "-druck.html", 1))
	})

	coll.OnHTML("html", func(elem *colly.HTMLElement) {
		var sb strings.Builder
		elem.ForEach("p", func(_ int, el *colly.HTMLElement) {
			p, err := html2text.FromString(el.Text)
			if err != nil {
				logrus.Warnf("could not paragraph from article: %s", err)
				return
			}
			sb.WriteString(p)
		})

		content := sb.String()
		createdAt := time.Now().Unix()
		url := elem.Request.URL.String()

		articles = append(articles, store.Article{
			ID:          fmt.Sprintf("%s-%d", url, createdAt),
			URL:         url,
			Headline:    elem.ChildText(".headline"),
			Description: meta(elem, "description"),
			Keywords:    keywords(elem),
			Content:     content,
			CreatedAt:   createdAt,
		})
	})

	q.Run(coll)

	return articles, nil
}

func meta(elem *colly.HTMLElement, name string) string {
	return elem.ChildAttr(fmt.Sprintf(`meta[name="%s"]`, name), "content")
}

func keywords(elem *colly.HTMLElement) []string {
	rawKeywords := strings.Split(meta(elem, "news_keywords"), ",")

	keywords := make([]string, 0, len(rawKeywords))
	for _, keyword := range rawKeywords {
		if k := strings.Trim(keyword, " "); k != "" {
			keywords = append(keywords, k)
		}
	}

	return keywords
}
