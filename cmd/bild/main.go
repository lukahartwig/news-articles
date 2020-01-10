package main

import (
	"os"
	"time"

	"github.com/gocolly/redisstorage"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/lukahartwig/news-articles/scraper"
	"github.com/lukahartwig/news-articles/store"
)

type config struct {
	MongoAddr   string
	RedisAddr   string
	RedisExpire time.Duration
	Topics      []string
}

func main() {
	app := &cli.App{
		Name: "bild-scraper",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "redis-addr",
				Value:   "localhost:6379",
				Usage:   "address of redis storage",
				EnvVars: []string{"REDIS_ADDR", "REDIS_URL"},
			},
			&cli.DurationFlag{
				Name:    "redis-expire",
				Value:   15 * time.Second,
				Usage:   "revisit timeout for urls",
				EnvVars: []string{"REDIS_EXPIRE"},
			},
			&cli.StringFlag{
				Name:    "mongo-addr",
				Value:   "mongodb://localhost:27017",
				Usage:   "address of mongo storage",
				EnvVars: []string{"MONGO_ADDR", "MONGO_URL"},
			},
		},
		Action: func(c *cli.Context) error {
			config := config{
				MongoAddr:   c.String("mongo-addr"),
				RedisAddr:   c.String("redis-addr"),
				RedisExpire: c.Duration("redis-expire"),
			}

			storage := store.New(config.MongoAddr)

			queueStorage := &redisstorage.Storage{
				Address: config.RedisAddr,
				Prefix:  "bild",
				Expires: config.RedisExpire,
			}
			defer func() {
				queueStorage.Client.Close()
			}()

			scraper := scraper.Scraper{
				Storage:   queueStorage,
				Extractor: scraper.BildExtractor{},
			}

			for {
				articles := scraper.Scrape("https://www.bild.de/rssfeeds/vw-alles/vw-alles-26970192,dzbildplus=false,sort=1,teaserbildmobil=false,view=rss2,wtmc=ob.feed.bild.xml")
				if len(articles) > 0 {
					logrus.Infof("saving %d articles", len(articles))
					storage.Save(articles)
				}
				time.Sleep(15 * time.Minute)
			}
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
