package core

import "time"

type Story struct {
	ID          string
	TopicID     string
	SourceID    string
	SourceName  string
	Title       string
	URL         string
	Author      string
	PublishedAt time.Time
	FetchedAt   time.Time
	Excerpt     string
	Content     string
	Score       float64
	Tags        []string
}
