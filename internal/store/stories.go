package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"github.com/patjrobinson/foxden/internal/core"
)

func (db *DB) UpsertStories(ctx context.Context, stories []core.Story) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin story upsert transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO stories (
    id,
    topic_id,
    source_id,
    source_name,
    title,
    url,
    author,
    published_at,
    fetched_at,
    excerpt,
    content,
    score,
    tags_json,
		content_source,
		article_fetched_at,
		article_fetch_error
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(topic_id, url) DO UPDATE SET
    source_id = excluded.source_id,
    source_name = excluded.source_name,
    title = excluded.title,
    author = excluded.author,
    published_at = excluded.published_at,
    fetched_at = excluded.fetched_at,
    excerpt = excluded.excerpt,
    content = excluded.content,
    score = excluded.score,
    tags_json = excluded.tags_json;
		content_source = excluded.content_source,
		article_fetched_at = excluded.article_fetched_at,
		article_fetch_error = excluded.article_fetch_error
`)
	if err != nil {
		return fmt.Errorf("prepare story upsert: %w", err)
	}
	defer stmt.Close()

	for _, story := range stories {
		tagsJSON, err := json.Marshal(story.Tags)
		if err != nil {
			return fmt.Errorf("marshal tags for story %q: %w", story.ID, err)
		}

		if _, err := stmt.ExecContext(
			ctx,
			story.ID,
			story.TopicID,
			story.SourceID,
			story.SourceName,
			story.Title,
			story.URL,
			story.Author,
			formatTime(story.PublishedAt),
			formatTime(story.FetchedAt),
			story.Excerpt,
			story.Content,
			story.Score,
			string(tagsJSON),
			contentSource(story.ContentSource),
			formatTime(story.ArticleFetchedAt),
			story.ArticleFetchError,
		); err != nil {
			return fmt.Errorf("upsert story %q: %w", story.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit story upsert transaction: %w", err)
	}

	return nil
}

func contentSource(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "feed"
	}
	return value
}

func (db *DB) StoriesForTopic(ctx context.Context, topicID string, since time.Time) ([]core.Story, error) {
	rows, err := db.sql.QueryContext(ctx, `
SELECT
    id,
    topic_id,
    source_id,
    source_name,
    title,
    url,
    author,
    published_at,
    fetched_at,
    excerpt,
    content,
    score,
    tags_json,
		content_source,
		article_fetched_at,
		article_fetch_error
FROM stories
WHERE topic_id = ?
  AND (
    ? IS NULL
    OR published_at IS NULL
    OR published_at >= ?
  )
ORDER BY
    COALESCE(published_at, fetched_at) DESC,
    fetched_at DESC
`, topicID, nullableTime(since), nullableTime(since))
	if err != nil {
		return nil, fmt.Errorf("query stories for topic %q: %w", topicID, err)
	}
	defer rows.Close()

	var stories []core.Story

	for rows.Next() {
		story, err := scanStory(rows)
		if err != nil {
			return nil, err
		}

		stories = append(stories, story)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stories for topic %q: %w", topicID, err)
	}

	return stories, nil
}

type storyScanner interface {
	Scan(dest ...any) error
}

func scanStory(row storyScanner) (core.Story, error) {
	var story core.Story
	var publishedAt sql.NullString
	var fetchedAt string
	var tagsJSON sql.NullString
	var contentSource sql.NullString
	var articleFetchedAt sql.NullString
	var articleFetchError sql.NullString

	if err := row.Scan(
		&story.ID,
		&story.TopicID,
		&story.SourceID,
		&story.SourceName,
		&story.Title,
		&story.URL,
		&story.Author,
		&publishedAt,
		&fetchedAt,
		&story.Excerpt,
		&story.Content,
		&story.Score,
		&tagsJSON,
		&contentSource,
		&articleFetchedAt,
		&articleFetchError
	); err != nil {
		return core.Story{}, fmt.Errorf("scan story: %w", err)
	}

	if publishedAt.Valid && publishedAt.String != "" {
		parsed, err := time.Parse(time.RFC3339, publishedAt.String)
		if err != nil {
			return core.Story{}, fmt.Errorf("parse published_at for story %q: %w", story.ID, err)
		}

		story.PublishedAt = parsed
	}

	if fetchedAt != "" {
		parsed, err := time.Parse(time.RFC3339, fetchedAt)
		if err != nil {
			return core.Story{}, fmt.Errorf("parse fetched_at for story %q: %w", story.ID, err)
		}

		story.FetchedAt = parsed
	}

	if tagsJSON.Valid && tagsJSON.String != "" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &story.Tags); err != nil {
			return core.Story{}, fmt.Errorf("parse tags for story %q: %w", story.ID, err)
		}
	}

	if contentSource.Valid {
		story.ContentSource = contentSource.String
	}

	if articleFetchedAt.Valid && articleFetchedAt.String != "" {
		parsed, err := time.Parse(time.RFC3339, articleFetchedAt.String)
		if err != nil {
			return core.Story{}, fmt.Errorf("parse article_fetched_at for story %q: %w", story.ID, err)
		}
		story.ArticleFetchedAt = parsed
	}

	if articleFetchError.Valid {
		story.ArticleFetchError = articleFetchError.String
	}

	return story, nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}

	return t.UTC().Format(time.RFC3339)
}

func formatTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}

	return t.UTC().Format(time.RFC3339)
}
