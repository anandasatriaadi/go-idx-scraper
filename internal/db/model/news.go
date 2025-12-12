package model

//go:generate mongogen -type=News -collection=news
type News struct {
	Title   string `bson:"title"`
	Date    string `bson:"date"`
	Summary string `bson:"summary"`
	Source  string `bson:"source"`
}
