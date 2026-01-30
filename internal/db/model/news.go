package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//go:generate go run ../../../generator/mongogen/main.go -type=News -collection=news
type News struct {
	Id       bson.ObjectID `bson:"_id"`
	Title    string        `bson:"title"`
	Date     time.Time     `bson:"date"`
	Summary  string        `bson:"summary"`
	Content  string        `bson:"content"`
	Priority int           `bson:"priority"`
	Link     string        `bson:"link"`
}
