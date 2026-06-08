package app

import (
	"context"
	"time"

	"github.com/patjrobinson/news-tui/internal/core"
	"github.com/patjrobinson/news-tui/internal/ingest"
	"github.com/patjrobinson/news-tui/internal/store"
)

type Services struct {
	Store  *store.DB
	Ingest ingest.Manager
	Now    func() time.Time
}

func NewServices(db *store.DB) Services {
	return Services{
		Store: db,
		Ingest: ingest.NewManager(map[string]ingest.Fetcher{
			"rss":             ingest.NewRSSFetcher(),
			"github_releases": ingest.NewGitHubReleasesFetcher(),
		}),
		Now: time.Now,
	}
}

func (s Services) RefreshTopic(ctx context.Context, topic core.Topic) ([]core.Story, error) {
	stories, err := s.Ingest.RefreshTopic(ctx, topic)
	if err != nil {
		return nil, err
	}

	if s.Store != nil && len(stories) > 0 {
		if err := s.Store.UpsertStories(ctx, stories); err != nil {
			return nil, err
		}
	}

	return stories, nil
}

func (s Services) LoadStories(ctx context.Context, topic core.Topic, r core.TimeRange) ([]core.Story, error) {
	if s.Store == nil {
		return nil, nil
	}

	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}

	return s.Store.StoriesForTopic(ctx, topic.ID, r.Since(now))
}
