package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/patjrobinson/news-tui/internal/core"
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
    tags_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		); err != nil {
			return fmt.Errorf("upsert story %q: %w", story.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit story upsert transaction: %w", err)
	}

	return nil
}

func formatTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}

	return t.UTC().Format(time.RFC3339)
}
