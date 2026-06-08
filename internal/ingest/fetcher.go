package ingest

import (
	"context"

	"github.com/patjrobinson/foxden/internal/core"
)

type Fetcher interface {
	Fetch(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error)
}

type FetcherFunc func(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error)

func (f FetcherFunc) Fetch(ctx context.Context, topic core.Topic, source core.Source) ([]core.Story, error) {
	return f(ctx, topic, source)
}
