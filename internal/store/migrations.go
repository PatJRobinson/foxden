package store

const initialSchema = `
CREATE TABLE IF NOT EXISTS stories (
    id TEXT PRIMARY KEY,
    topic_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    source_name TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    author TEXT,
    published_at TEXT,
    fetched_at TEXT NOT NULL,
    excerpt TEXT,
    content TEXT,
    score REAL DEFAULT 0,
    tags_json TEXT,
		content_source TEXT DEFAULT 'feed',
		article_fetched_at TEXT,
		article_fetch_error TEXT,
    UNIQUE(topic_id, url)
);


CREATE TABLE IF NOT EXISTS summaries (
    id TEXT PRIMARY KEY,
    topic_id TEXT NOT NULL,
    time_range TEXT NOT NULL,
    agents_hash TEXT,
    newest_story_time TEXT,
    generated_at TEXT NOT NULL,
    content TEXT NOT NULL
);
`
const storyArticleColumnsMigration = `
ALTER TABLE stories ADD COLUMN content_source TEXT DEFAULT 'feed';
ALTER TABLE stories ADD COLUMN article_fetched_at TEXT;
ALTER TABLE stories ADD COLUMN article_fetch_error TEXT;
`
