package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgraph-io/ristretto"
	"github.com/zaker/anachrome-be/stores/blog"
)

type BlogCache struct {
	persist blog.BlogStore
	cache   *ristretto.Cache
}

type CacheError error

func NewBlogCache(p blog.BlogStore) (*BlogCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("Creating new blog cache: %w", err)
	}
	return &BlogCache{persist: p, cache: cache}, nil
}

func (bc *BlogCache) GetBlogPost(ctx context.Context, id string) (blog.BlogPost, error) {

	return bc.persist.GetBlogPost(ctx, id)
}

func (bc *BlogCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {

	v, ok := bc.cache.Get("PostsMeta")

	if ok {
		return v.([]blog.BlogPostMeta), nil
	}

	bpm, err := bc.persist.GetBlogPostsMeta(ctx)
	if err != nil {
		return nil, err
	}
	ok = bc.cache.Set("PostsMeta", bpm, 0)
	if !ok {
		return nil, CacheError(errors.New("Failed to add to cache"))
	}
	return bpm, nil
}
