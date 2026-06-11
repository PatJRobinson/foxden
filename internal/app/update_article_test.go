package app

import (
	"errors"
	"testing"
	"time"

	"github.com/patjrobinson/foxden/internal/core"
)

func TestUpdateArticleFetchedUpdatesStoryListAndReader(t *testing.T) {
	oldStory := core.Story{
		ID:            "story-1",
		TopicID:       "topic-1",
		Title:         "Old",
		Content:       "feed content",
		ContentSource: "feed",
	}

	articleFetchedAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	updatedStory := oldStory
	updatedStory.Content = "article content"
	updatedStory.ContentSource = "article"
	updatedStory.ArticleFetchedAt = articleFetchedAt

	model := Model{
		Stories: map[string][]core.Story{
			"topic-1": {oldStory},
		},
		ViewMode:        ViewModeReader,
		ReaderStory:     oldStory,
		ReaderOffset:    10,
		FetchingArticle: true,
		Err:             errors.New("old error"),
	}

	gotModel, cmd := model.Update(articleFetchedMsg{Story: updatedStory})
	if cmd != nil {
		t.Fatalf("Update() cmd = %v, want nil", cmd)
	}

	got := gotModel.(Model)

	if got.FetchingArticle {
		t.Fatal("FetchingArticle = true, want false")
	}

	if got.Err != nil {
		t.Fatalf("Err = %v, want nil", got.Err)
	}

	if got.ReaderStory.Content != "article content" {
		t.Fatalf("ReaderStory.Content = %q, want article content", got.ReaderStory.Content)
	}

	if got.ReaderStory.ContentSource != "article" {
		t.Fatalf("ReaderStory.ContentSource = %q, want article", got.ReaderStory.ContentSource)
	}

	if !got.ReaderStory.ArticleFetchedAt.Equal(articleFetchedAt) {
		t.Fatalf("ReaderStory.ArticleFetchedAt = %v, want %v", got.ReaderStory.ArticleFetchedAt, articleFetchedAt)
	}

	if got.ReaderOffset != 0 {
		t.Fatalf("ReaderOffset = %d, want 0", got.ReaderOffset)
	}

	stories := got.Stories["topic-1"]
	if len(stories) != 1 {
		t.Fatalf("len(stories) = %d, want 1", len(stories))
	}

	if stories[0].Content != "article content" {
		t.Fatalf("stored Content = %q, want article content", stories[0].Content)
	}

	if stories[0].ContentSource != "article" {
		t.Fatalf("stored ContentSource = %q, want article", stories[0].ContentSource)
	}
}

func TestUpdateArticleFetchFailedSetsError(t *testing.T) {
	fetchErr := errors.New("fetch failed")

	model := Model{
		FetchingArticle: true,
	}

	gotModel, cmd := model.Update(articleFetchFailedMsg{
		StoryID: "story-1",
		Err:     fetchErr,
	})
	if cmd != nil {
		t.Fatalf("Update() cmd = %v, want nil", cmd)
	}

	got := gotModel.(Model)

	if got.FetchingArticle {
		t.Fatal("FetchingArticle = true, want false")
	}

	if got.Err == nil || got.Err.Error() != "fetch failed" {
		t.Fatalf("Err = %v, want fetch failed", got.Err)
	}
}

func TestUpdateStoryReplacesMatchingStory(t *testing.T) {
	oldStory := core.Story{
		ID:      "story-1",
		TopicID: "topic-1",
		Title:   "Old",
	}

	model := Model{
		Stories: map[string][]core.Story{
			"topic-1": {
				oldStory,
				{ID: "story-2", TopicID: "topic-1", Title: "Other"},
			},
		},
	}

	updated := oldStory
	updated.Title = "Updated"

	model.updateStory(updated)

	if got := model.Stories["topic-1"][0].Title; got != "Updated" {
		t.Fatalf("updated story title = %q, want Updated", got)
	}

	if got := model.Stories["topic-1"][1].Title; got != "Other" {
		t.Fatalf("unrelated story title = %q, want Other", got)
	}
}
