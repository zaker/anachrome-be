package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"encoding/gob"

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

	b := bytes.Buffer{}
	gob.Register(blog.BlogPost{})

	m, ok := imbc.cache.Get(id)
	if ok {
		dec := gob.NewDecoder(&b)
		bp := blog.BlogPost{}
		b.Write(m)
		err := dec.Decode(&bp)
		if err != nil {
			return bp, fmt.Errorf("decoding cached item")
		}
		return bp, nil
	}

	enc := gob.NewEncoder(&b)
	bp, err := imbc.persist.GetBlogPost(ctx, id)
	if err != nil {
		return blog.BlogPost{}, err
	}
	err = enc.Encode(&bp)

	if err != nil {
		return blog.BlogPost{}, err
	}

	m, err = io.ReadAll(&b)
	if err != nil {
		return blog.BlogPost{}, err
	}

	imbc.cache.Set(id, m)
	return bp, nil
}

func (imbc *InMemoryCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {
	return imbc.persist.GetBlogPostsMeta(ctx)
}

func (imbc *InMemoryCache) Invalidate(ctx context.Context, id string) error {
	imbc.cache.Del(id)
	return nil
}
