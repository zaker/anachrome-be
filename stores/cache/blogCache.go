package cache

import (
	"context"

	"github.com/zaker/anachrome-be/stores/blog"
)

type CacheError error

type CachedBlogStore interface {
	blog.BlogStore
	Invalidate(context.Context, string) error
}
