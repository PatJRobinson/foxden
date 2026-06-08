package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/patjrobinson/news-tui/internal/core"
)

func TestManagerRefreshTopic(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	manager := NewManager(map[string]Fetcher{
		"rss": FetcherFunc(func(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error) {
			return []core.Story{
				{
					ID:    "story-1",
					Title: "Test story",
					URL:   "https://example.com/story",
				},
			}, nil
		}),
	})
	manager.Now = func() time.Time {
		return now
	}

	topic := core.Topic{
		ID:    "test-topic",
		Title: "Test Topic",
		Sources: []core.Source{
			{
				ID:   "test-source",
				Name: "Test Source",
				Type: "rss",
				URL:  "https://example.com/rss",
			},
		},
	}

	stories, err := manager.RefreshTopic(context.Background(), topic)
	if err != nil {
		t.Fatalf("RefreshTopic() error = %v", err)
	}

	if len(stories) != 1 {
		t.Fatalf("RefreshTopic() returned %d stories, want 1", len(stories))
	}

	story := stories[0]

	if story.TopicID != "test-topic" {
		t.Fatalf("story.TopicID = %q, want %q", story.TopicID, "test-topic")
	}

	if story.SourceID != "test-source" {
		t.Fatalf("story.SourceID = %q, want %q", story.SourceID, "test-source")
	}

	if story.SourceName != "Test Source" {
		t.Fatalf("story.SourceName = %q, want %q", story.SourceName, "Test Source")
	}

	if !story.FetchedAt.Equal(now) {
		t.Fatalf("story.FetchedAt = %v, want %v", story.FetchedAt, now)
	}
}

func TestManagerRefreshTopicMissingFetcher(t *testing.T) {
	manager := NewManager(nil)

	topic := core.Topic{
		ID:    "test-topic",
		Title: "Test Topic",
		Sources: []core.Source{
			{
				ID:   "test-source",
				Name: "Test Source",
				Type: "rss",
			},
		},
	}

	_, err := manager.RefreshTopic(context.Background(), topic)
	if err == nil {
		t.Fatal("RefreshTopic() error = nil, want error")
	}

	if !strings.Contains(err.Error(), `no fetcher registered for source type "rss"`) {
		t.Fatalf("RefreshTopic() error = %q", err)
	}
}

func TestManagerRefreshTopicFetcherError(t *testing.T) {
	manager := NewManager(map[string]Fetcher{
		"rss": FetcherFunc(func(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error) {
			return nil, errors.New("boom")
		}),
	})

	topic := core.Topic{
		ID:    "test-topic",
		Title: "Test Topic",
		Sources: []core.Source{
			{
				ID:   "test-source",
				Name: "Test Source",
				Type: "rss",
			},
		},
	}

	_, err := manager.RefreshTopic(context.Background(), topic)
	if err == nil {
		t.Fatal("RefreshTopic() error = nil, want error")
	}

	if !strings.Contains(err.Error(), `fetch source "test-source": boom`) {
		t.Fatalf("RefreshTopic() error = %q", err)
	}
}
