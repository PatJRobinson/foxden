package store

import (
	"context"
	"testing"
	"time"

	"github.com/patjrobinson/foxden/internal/core"
)

func TestDBInitAndUpsertStories(t *testing.T) {
	ctx := context.Background()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	story := core.Story{
		ID:          "story-1",
		TopicID:     "topic-1",
		SourceID:    "source-1",
		SourceName:  "Source One",
		Title:       "Original title",
		URL:         "https://example.com/story",
		Author:      "Test Author",
		PublishedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC),
		FetchedAt:   time.Date(2026, 6, 8, 11, 0, 0, 0, time.UTC),
		Excerpt:     "Original excerpt",
		Content:     "Original content",
		Score:       0.5,
		Tags:        []string{"go", "tui"},
	}

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() insert error = %v", err)
	}

	var title string
	err = db.sql.QueryRowContext(ctx, `
SELECT title FROM stories WHERE topic_id = ? AND url = ?
`, story.TopicID, story.URL).Scan(&title)
	if err != nil {
		t.Fatalf("query inserted story: %v", err)
	}

	if title != "Original title" {
		t.Fatalf("inserted title = %q, want %q", title, "Original title")
	}

	story.ID = "story-1-updated"
	story.Title = "Updated title"
	story.Excerpt = "Updated excerpt"

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() update error = %v", err)
	}

	var updatedTitle string
	var updatedExcerpt string

	err = db.sql.QueryRowContext(ctx, `
SELECT title, excerpt FROM stories WHERE topic_id = ? AND url = ?
`, story.TopicID, story.URL).Scan(&updatedTitle, &updatedExcerpt)
	if err != nil {
		t.Fatalf("query updated story: %v", err)
	}

	if updatedTitle != "Updated title" {
		t.Fatalf("updated title = %q, want %q", updatedTitle, "Updated title")
	}

	if updatedExcerpt != "Updated excerpt" {
		t.Fatalf("updated excerpt = %q, want %q", updatedExcerpt, "Updated excerpt")
	}
}

func TestStoriesForTopicFiltersAndSorts(t *testing.T) {
	ctx := context.Background()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	oldStory := core.Story{
		ID:          "old-story",
		TopicID:     "topic-1",
		SourceID:    "source-1",
		SourceName:  "Source One",
		Title:       "Old story",
		URL:         "https://example.com/old",
		PublishedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		FetchedAt:   time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC),
	}

	newStory := core.Story{
		ID:          "new-story",
		TopicID:     "topic-1",
		SourceID:    "source-1",
		SourceName:  "Source One",
		Title:       "New story",
		URL:         "https://example.com/new",
		PublishedAt: time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC),
		FetchedAt:   time.Date(2026, 6, 8, 10, 5, 0, 0, time.UTC),
	}

	otherTopicStory := core.Story{
		ID:          "other-topic-story",
		TopicID:     "topic-2",
		SourceID:    "source-1",
		SourceName:  "Source One",
		Title:       "Other topic story",
		URL:         "https://example.com/other",
		PublishedAt: time.Date(2026, 6, 8, 11, 0, 0, 0, time.UTC),
		FetchedAt:   time.Date(2026, 6, 8, 11, 5, 0, 0, time.UTC),
	}

	if err := db.UpsertStories(ctx, []core.Story{oldStory, newStory, otherTopicStory}); err != nil {
		t.Fatalf("UpsertStories() error = %v", err)
	}

	since := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)

	stories, err := db.StoriesForTopic(ctx, "topic-1", since)
	if err != nil {
		t.Fatalf("StoriesForTopic() error = %v", err)
	}

	if len(stories) != 1 {
		t.Fatalf("StoriesForTopic() returned %d stories, want 1", len(stories))
	}

	if stories[0].ID != "new-story" {
		t.Fatalf("StoriesForTopic()[0].ID = %q, want %q", stories[0].ID, "new-story")
	}
}
