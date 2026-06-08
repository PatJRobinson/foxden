package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/patjrobinson/news-tui/internal/core"
)

type RSSFetcher struct {
	Client *http.Client
	Parser *gofeed.Parser
}

func NewRSSFetcher() RSSFetcher {
	return RSSFetcher{
		Client: &http.Client{Timeout: 20 * time.Second},
		Parser: gofeed.NewParser(),
	}
}

func (f RSSFetcher) Fetch(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error) {
	if strings.TrimSpace(source.URL) == "" {
		return nil, fmt.Errorf("rss source %q is missing url", source.ID)
	}

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	parser := f.Parser
	if parser == nil {
		parser = gofeed.NewParser()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create rss request %q: %w", source.URL, err)
	}

	req.Header.Set("User-Agent", "news-tui/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rss url %q: %w", source.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch rss url %q: unexpected status %s", source.URL, resp.Status)
	}

	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse rss url %q: %w", source.URL, err)
	}

	stories := make([]core.Story, 0, len(feed.Items))
	now := time.Now()

	for _, item := range feed.Items {
		url := strings.TrimSpace(item.Link)
		if url == "" {
			url = strings.TrimSpace(item.GUID)
		}

		title := strings.TrimSpace(item.Title)
		if title == "" || url == "" {
			continue
		}

		stories = append(stories, core.Story{
			ID:          storyID(topic.ID, source.ID, url),
			TopicID:     topic.ID,
			SourceID:    source.ID,
			SourceName:  source.Name,
			Title:       title,
			URL:         url,
			Author:      authorName(item),
			PublishedAt: publishedTime(item),
			FetchedAt:   now,
			Excerpt:     firstNonEmpty(item.Description, item.Content),
			Content:     item.Content,
		})
	}

	return stories, nil
}

func publishedTime(item *gofeed.Item) time.Time {
	if item.PublishedParsed != nil {
		return *item.PublishedParsed
	}

	if item.UpdatedParsed != nil {
		return *item.UpdatedParsed
	}

	return time.Time{}
}

func authorName(item *gofeed.Item) string {
	if item.Author != nil {
		return item.Author.Name
	}

	if len(item.Authors) > 0 {
		return item.Authors[0].Name
	}

	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func storyID(topicID string, sourceID string, url string) string {
	sum := sha256.Sum256([]byte(topicID + "\x00" + sourceID + "\x00" + url))
	return hex.EncodeToString(sum[:])
}
