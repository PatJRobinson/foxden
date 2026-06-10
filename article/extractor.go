package article

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"
)

type ExtractedArticle struct {
	Title       string
	Byline      string
	SiteName    string
	TextContent string
	Excerpt     string
	Length      int
	FetchedAt   time.Time
}

type Extractor struct {
	Client *http.Client
	Now    func() time.Time
}

func NewExtractor() Extractor {
	return Extractor{
		Client: &http.Client{Timeout: 25 * time.Second},
		Now:    time.Now,
	}
}

func (e Extractor) Extract(ctx context.Context, rawURL string) (ExtractedArticle, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ExtractedArticle{}, fmt.Errorf("missing article url")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ExtractedArticle{}, fmt.Errorf("parse article url %q: %w", rawURL, err)
	}

	client := e.Client
	if client == nil {
		client = &http.Client{Timeout: 25 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ExtractedArticle{}, fmt.Errorf("create article request %q: %w", rawURL, err)
	}

	req.Header.Set("User-Agent", "foxden/0.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return ExtractedArticle{}, fmt.Errorf("fetch article %q: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ExtractedArticle{}, fmt.Errorf("fetch article %q: unexpected status %s", rawURL, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "text/html") {
		return ExtractedArticle{}, fmt.Errorf("fetch article %q: unsupported content type %q", rawURL, contentType)
	}

	parsedArticle, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return ExtractedArticle{}, fmt.Errorf("extract readable article %q: %w", rawURL, err)
	}

	text := strings.TrimSpace(parsedArticle.TextContent)
	if text == "" {
		return ExtractedArticle{}, fmt.Errorf("extract readable article %q: no readable text found", rawURL)
	}

	now := time.Now()
	if e.Now != nil {
		now = e.Now()
	}

	return ExtractedArticle{
		Title:       strings.TrimSpace(parsedArticle.Title),
		Byline:      strings.TrimSpace(parsedArticle.Byline),
		SiteName:    strings.TrimSpace(parsedArticle.SiteName),
		TextContent: text,
		Excerpt:     strings.TrimSpace(parsedArticle.Excerpt),
		Length:      parsedArticle.Length,
		FetchedAt:   now,
	}, nil
}
