package article

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExtractorExtractsReadableArticle(t *testing.T) {
	fetchedAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "foxden/0.1" {
			t.Fatalf("User-Agent = %q, want foxden/0.1", got)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html>
<head>
  <title>Test Article</title>
</head>
<body>
  <nav>ignore this navigation</nav>
  <article>
    <h1>Test Article</h1>
    <p>This is the first useful paragraph of the article.</p>
    <p>This is the second useful paragraph with enough text to be readable.</p>
  </article>
</body>
</html>`))
	}))
	defer server.Close()

	extractor := Extractor{
		Client: server.Client(),
		Now: func() time.Time {
			return fetchedAt
		},
	}

	article, err := extractor.Extract(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if !article.FetchedAt.Equal(fetchedAt) {
		t.Fatalf("FetchedAt = %v, want %v", article.FetchedAt, fetchedAt)
	}

	if !strings.Contains(article.TextContent, "first useful paragraph") {
		t.Fatalf("TextContent did not contain article body: %q", article.TextContent)
	}
}

func TestExtractorRejectsMissingURL(t *testing.T) {
	extractor := NewExtractor()

	_, err := extractor.Extract(context.Background(), " ")
	if err == nil {
		t.Fatal("Extract() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "missing article url") {
		t.Fatalf("Extract() error = %q, want missing article url", err.Error())
	}
}

func TestExtractorRejectsNonHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	extractor := Extractor{
		Client: server.Client(),
	}

	_, err := extractor.Extract(context.Background(), server.URL)
	if err == nil {
		t.Fatal("Extract() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unsupported content type") {
		t.Fatalf("Extract() error = %q, want unsupported content type", err.Error())
	}
}

func TestExtractorRejectsBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	extractor := Extractor{
		Client: server.Client(),
	}

	_, err := extractor.Extract(context.Background(), server.URL)
	if err == nil {
		t.Fatal("Extract() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("Extract() error = %q, want unexpected status", err.Error())
	}
}
