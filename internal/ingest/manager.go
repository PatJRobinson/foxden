package ingest

import (
	"context"
	"fmt"
	"time"

	"github.com/patjrobinson/news-tui/internal/core"
)

type SourceResult struct {
	Source  core.Source
	Stories []core.Story
	Err     error
}

type Manager struct {
	Fetchers map[string]Fetcher
	Now      func() time.Time
}

func NewManager(fetchers map[string]Fetcher) Manager {
	return Manager{
		Fetchers: fetchers,
		Now:      time.Now,
	}
}

func (m Manager) RefreshTopic(ctx context.Context, topic core.Topic) ([]core.Story, error) {
	var all []core.Story

	for _, source := range topic.Sources {
		fetcher, ok := m.Fetchers[source.Type]
		if !ok {
			return nil, fmt.Errorf("no fetcher registered for source type %q", source.Type)
		}

		stories, err := fetcher.Fetch(ctx, topic, source)
		if err != nil {
			return nil, fmt.Errorf("fetch source %q: %w", source.ID, err)
		}

		fetchedAt := m.now()

		for i := range stories {
			if stories[i].TopicID == "" {
				stories[i].TopicID = topic.ID
			}

			if stories[i].SourceID == "" {
				stories[i].SourceID = source.ID
			}

			if stories[i].SourceName == "" {
				stories[i].SourceName = source.Name
			}

			if stories[i].FetchedAt.IsZero() {
				stories[i].FetchedAt = fetchedAt
			}
		}

		all = append(all, stories...)
	}

	return all, nil
}

func (m Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}

	return time.Now()
}
