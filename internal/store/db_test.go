package store

import (
	"context"
	"database/sql"
	"errors"
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

	articleFetchedAt := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	story := core.Story{
		ID:                "story-1",
		TopicID:           "topic-1",
		SourceID:          "source-1",
		SourceName:        "Source One",
		Title:             "Original title",
		URL:               "https://example.com/story",
		Author:            "Test Author",
		PublishedAt:       time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC),
		FetchedAt:         time.Date(2026, 6, 8, 11, 0, 0, 0, time.UTC),
		Excerpt:           "Original excerpt",
		Content:           "Original content",
		Score:             0.5,
		Tags:              []string{"go", "tui"},
		ContentSource:     "article",
		ArticleFetchedAt:  articleFetchedAt,
		ArticleFetchError: "old fetch error",
	}

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() insert error = %v", err)
	}

	var title string
	var contentSource string
	var articleFetchedAtText string
	var articleFetchError string

	err = db.sql.QueryRowContext(ctx, `
SELECT title, content_source, article_fetched_at, article_fetch_error
FROM stories
WHERE topic_id = ? AND url = ?
`, story.TopicID, story.URL).Scan(
		&title,
		&contentSource,
		&articleFetchedAtText,
		&articleFetchError,
	)
	if err != nil {
		t.Fatalf("query inserted story: %v", err)
	}

	if title != "Original title" {
		t.Fatalf("inserted title = %q, want %q", title, "Original title")
	}

	if contentSource != "article" {
		t.Fatalf("inserted content_source = %q, want %q", contentSource, "article")
	}

	if articleFetchedAtText != articleFetchedAt.Format(time.RFC3339) {
		t.Fatalf("inserted article_fetched_at = %q, want %q", articleFetchedAtText, articleFetchedAt.Format(time.RFC3339))
	}

	if articleFetchError != "old fetch error" {
		t.Fatalf("inserted article_fetch_error = %q, want %q", articleFetchError, "old fetch error")
	}

	story.ID = "story-1-updated"
	story.Title = "Updated title"
	story.Excerpt = "Updated excerpt"
	story.ContentSource = "feed"
	story.ArticleFetchedAt = time.Time{}
	story.ArticleFetchError = ""

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() update error = %v", err)
	}

	var updatedTitle string
	var updatedExcerpt string
	var updatedContentSource string
	var updatedArticleFetchedAt sql.NullString
	var updatedArticleFetchError sql.NullString

	err = db.sql.QueryRowContext(ctx, `
SELECT title, excerpt, content_source, article_fetched_at, article_fetch_error
FROM stories
WHERE topic_id = ? AND url = ?
`, story.TopicID, story.URL).Scan(
		&updatedTitle,
		&updatedExcerpt,
		&updatedContentSource,
		&updatedArticleFetchedAt,
		&updatedArticleFetchError,
	)
	if err != nil {
		t.Fatalf("query updated story: %v", err)
	}

	if updatedTitle != "Updated title" {
		t.Fatalf("updated title = %q, want %q", updatedTitle, "Updated title")
	}

	if updatedExcerpt != "Updated excerpt" {
		t.Fatalf("updated excerpt = %q, want %q", updatedExcerpt, "Updated excerpt")
	}

	if updatedContentSource != "feed" {
		t.Fatalf("updated content_source = %q, want %q", updatedContentSource, "feed")
	}

	if updatedArticleFetchedAt.Valid {
		t.Fatalf("updated article_fetched_at = %q, want NULL", updatedArticleFetchedAt.String)
	}

	if updatedArticleFetchError.Valid {
		t.Fatalf("updated article_fetch_error = %q, want NULL", updatedArticleFetchError.String)
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

func TestUpdateStoryArticle(t *testing.T) {
	ctx := context.Background()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	fetchedAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	story := core.Story{
		ID:         "story-1",
		TopicID:    "topic-1",
		SourceID:   "source-1",
		SourceName: "Source",
		Title:      "Story",
		URL:        "https://example.com/story",
		FetchedAt:  fetchedAt.Add(-time.Hour),
		Content:    "feed content",
		Excerpt:    "feed excerpt",
	}

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() error = %v", err)
	}

	articleFetchedAt := time.Date(2026, 6, 10, 13, 0, 0, 0, time.UTC)

	if err := db.UpdateStoryArticle(ctx, "story-1", "article content", "article excerpt", articleFetchedAt); err != nil {
		t.Fatalf("UpdateStoryArticle() error = %v", err)
	}

	stories, err := db.StoriesForTopic(ctx, "topic-1", time.Time{})
	if err != nil {
		t.Fatalf("StoriesForTopic() error = %v", err)
	}

	if len(stories) != 1 {
		t.Fatalf("len(stories) = %d, want 1", len(stories))
	}

	got := stories[0]

	if got.Content != "article content" {
		t.Fatalf("Content = %q, want article content", got.Content)
	}

	if got.Excerpt != "article excerpt" {
		t.Fatalf("Excerpt = %q, want article excerpt", got.Excerpt)
	}

	if got.ContentSource != "article" {
		t.Fatalf("ContentSource = %q, want article", got.ContentSource)
	}

	if !got.ArticleFetchedAt.Equal(articleFetchedAt) {
		t.Fatalf("ArticleFetchedAt = %v, want %v", got.ArticleFetchedAt, articleFetchedAt)
	}

	if got.ArticleFetchError != "" {
		t.Fatalf("ArticleFetchError = %q, want empty", got.ArticleFetchError)
	}
}

func TestMarkStoryArticleFetchError(t *testing.T) {
	ctx := context.Background()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	story := core.Story{
		ID:         "story-1",
		TopicID:    "topic-1",
		SourceID:   "source-1",
		SourceName: "Source",
		Title:      "Story",
		URL:        "https://example.com/story",
		FetchedAt:  now,
	}

	if err := db.UpsertStories(ctx, []core.Story{story}); err != nil {
		t.Fatalf("UpsertStories() error = %v", err)
	}

	fetchedAt := now.Add(time.Hour)
	fetchErr := errors.New("server exploded")

	if err := db.MarkStoryArticleFetchError(ctx, "story-1", fetchedAt, fetchErr); err != nil {
		t.Fatalf("MarkStoryArticleFetchError() error = %v", err)
	}

	stories, err := db.StoriesForTopic(ctx, "topic-1", time.Time{})
	if err != nil {
		t.Fatalf("StoriesForTopic() error = %v", err)
	}

	if len(stories) != 1 {
		t.Fatalf("len(stories) = %d, want 1", len(stories))
	}

	got := stories[0]

	if !got.ArticleFetchedAt.Equal(fetchedAt) {
		t.Fatalf("ArticleFetchedAt = %v, want %v", got.ArticleFetchedAt, fetchedAt)
	}

	if got.ArticleFetchError != "server exploded" {
		t.Fatalf("ArticleFetchError = %q, want server exploded", got.ArticleFetchError)
	}

	if got.ContentSource != "feed" {
		t.Fatalf("ContentSource = %q, want feed", got.ContentSource)
	}
}
