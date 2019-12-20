package store

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Article struct {
	ID          string   `bson:"_id"`
	URL         string   `bson:"url"`
	Headline    string   `bson:"headline"`
	Description string   `bson:"description"`
	Keywords    []string `bson:"keywords"`
	Content     string   `bson:"content"`
	CreatedAt   int64    `bson:"createdAt"`
}

type Store interface {
	Save(articles []Article) error
}

type store struct {
	mongo *mongo.Client
}

func New(addr string) Store {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	if err != nil {
		logrus.Fatal(err)
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		logrus.Fatalf("mongodb not available: %v", err)
	}

	return &store{
		mongo: client,
	}
}

func (s *store) Save(articles []Article) error {
	documents := make([]mongo.WriteModel, len(articles))
	for i, article := range articles {
		filter := bson.D{
			{Key: "_id", Value: article.ID},
		}
		documents[i] = mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(article).SetUpsert(true)
	}

	result, err := s.mongo.Database("nosql").Collection("articles").BulkWrite(
		context.Background(),
		documents,
		options.BulkWrite().SetOrdered(false),
	)
	if err != nil {
		return fmt.Errorf("failed to insert article: %w", err)
	}

	logrus.Infof("%d inserted articles, %d updated articles, %d upserted articles", result.InsertedCount, result.ModifiedCount, result.UpsertedCount)

	return nil
}
