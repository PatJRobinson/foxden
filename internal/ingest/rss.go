package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
	"html"
	"io"
	"regexp"

	"github.com/mmcdole/gofeed"

	"github.com/patjrobinson/news-tui/internal/core"

	nethtml "golang.org/x/net/html"
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

		rawText := firstNonEmpty(item.Description, item.Content)
		cleanText := cleanFeedText(rawText)

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
			Excerpt:     truncateText(cleanText, 500),
			Content:     cleanText,
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

var (
	htmlTagRE        = regexp.MustCompile(`<[^>]+>`)
	whitespaceRE     = regexp.MustCompile(`\s+`)
	discourseQuoteRE = regexp.MustCompile(`(?is)<aside[^>]*class="quote"[^>]*>.*?</aside>`)
)

func cleanFeedText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var b strings.Builder
	tokenizer := nethtml.NewTokenizer(strings.NewReader(value))

	skipDepth := 0

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case nethtml.ErrorToken:
			if tokenizer.Err() == io.EOF {
				text := html.UnescapeString(b.String())
				text = whitespaceRE.ReplaceAllString(text, " ")
				return strings.TrimSpace(text)
			}

			// Fall back to a simple unescape if tokenization fails.
			text := html.UnescapeString(value)
			text = whitespaceRE.ReplaceAllString(text, " ")
			return strings.TrimSpace(text)

		case nethtml.StartTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)

			if tag == "script" || tag == "style" {
				skipDepth++
				continue
			}

			if tag == "aside" && hasClass(token, "quote") {
				skipDepth++
				continue
			}

			if skipDepth > 0 {
				continue
			}

			if tag == "p" || tag == "br" || tag == "li" || tag == "div" || tag == "blockquote" {
				b.WriteString(" ")
			}

		case nethtml.EndTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)

			if skipDepth > 0 {
				if tag == "script" || tag == "style" || tag == "aside" {
					skipDepth--
				}
				continue
			}

			if tag == "p" || tag == "li" || tag == "div" || tag == "blockquote" {
				b.WriteString(" ")
			}

		case nethtml.TextToken:
			if skipDepth > 0 {
				continue
			}

			text := strings.TrimSpace(string(tokenizer.Text()))
			if text != "" {
				b.WriteString(text)
				b.WriteString(" ")
			}
		}
	}
}

func hasClass(token nethtml.Token, className string) bool {
	for _, attr := range token.Attr {
		if strings.EqualFold(attr.Key, "class") {
			classes := strings.Fields(attr.Val)
			for _, class := range classes {
				if class == className {
					return true
				}
			}
		}
	}

	return false
}

func truncateText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}

	return strings.TrimSpace(value[:max]) + "…"
}
