package app

import (
	"context"
	"strings"
	"time"

	"github.com/patjrobinson/foxden/internal/article"
	"github.com/patjrobinson/foxden/internal/core"
	"github.com/patjrobinson/foxden/internal/ingest"
	"github.com/patjrobinson/foxden/internal/store"
)

type Services struct {
	Store   *store.DB
	Ingest  ingest.Manager
	Now     func() time.Time
	Article article.Extractor
}

func NewServices(db *store.DB) Services {
	return Services{
		Store: db,
		Ingest: ingest.NewManager(map[string]ingest.Fetcher{
			"rss":             ingest.NewRSSFetcher(),
			"github_releases": ingest.NewGitHubReleasesFetcher(),
		}),
		Now:     time.Now,
		Article: article.NewExtractor(),
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

func (s Services) FetchArticleForStory(ctx context.Context, story core.Story) (core.Story, error) {
	extracted, err := s.Article.Extract(ctx, story.URL)
	if err != nil {
		if s.Store != nil {
			_ = s.Store.MarkStoryArticleFetchError(ctx, story.ID, time.Now(), err)
		}
		return story, err
	}

	content := extracted.TextContent
	excerpt := extracted.Excerpt
	if strings.TrimSpace(excerpt) == "" {
		excerpt = truncateForService(content, 500)
	}

	if s.Store != nil {
		if err := s.Store.UpdateStoryArticle(ctx, story.ID, content, excerpt, extracted.FetchedAt); err != nil {
			return story, err
		}
	}

	story.Content = content
	story.Excerpt = excerpt
	story.ContentSource = "article"
	story.ArticleFetchedAt = extracted.FetchedAt
	story.ArticleFetchError = ""

	return story, nil
}

func truncateForService(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return strings.TrimSpace(value[:max]) + "…"
}
