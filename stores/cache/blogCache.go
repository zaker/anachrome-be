package cache

import (
	"context"

	"github.com/zaker/anachrome-be/stores/blog"
)

type BlogCache struct {
	persist blog.BlogStore
}

func NewBlogCache(p blog.BlogStore) *BlogCache {

	return &BlogCache{persist: p}
}

func (bc *BlogCache) GetBlogPost(ctx context.Context, id string) (blog.BlogPost, error) {

	return bc.persist.GetBlogPost(ctx, id)
}

func (bc *BlogCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {

	return bc.persist.GetBlogPostsMeta(ctx)
}
