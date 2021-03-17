package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/zaker/anachrome-be/stores/blog"
)

type InMemoryCache struct {
	persist blog.BlogStore

	cache *cache.TinyLFU
}

func NewInMemoryCache(p blog.BlogStore) (*InMemoryCache, error) {

	return &InMemoryCache{persist: p, cache: cache.NewTinyLFU(1000, time.Minute)}, nil
}

func (imbc *InMemoryCache) GetBlogPost(ctx context.Context, id string) (blog.BlogPost, error) {
	return imbc.persist.GetBlogPost(ctx, id)
}

func (imbc *InMemoryCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {
	return imbc.persist.GetBlogPostsMeta(ctx)
}

func (imbc *InMemoryCache) Invalidate(ctx context.Context, id string) error {

	return nil
}
