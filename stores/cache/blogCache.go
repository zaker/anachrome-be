package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/zaker/anachrome-be/stores/blog"
)

type BlogCache struct {
	persist blog.BlogStore
	cache   *cache.Cache
}

type CacheError error

func NewRedisBlogCache(p blog.BlogStore, redishost string) (*BlogCache, error) {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": redishost,
		},
	})

	cache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &BlogCache{persist: p, cache: cache}, nil
}

func (bc *BlogCache) GetBlogPost(ctx context.Context, id string) (blog.BlogPost, error) {

	var bp blog.BlogPost
	key := "post:" + id
	err := bc.cache.Get(ctx, key, &bp)

	if err == nil {
		return bp, nil
	}
	if err != nil && err != cache.ErrCacheMiss {
		return bp, CacheError(errors.New("Failed to get post{" + key + "} "))
	}

	bp, err = bc.persist.GetBlogPost(ctx, id)
	if err != nil {
		return bp, err
	}
	err = bc.cache.Set(
		&cache.Item{
			Ctx:   ctx,
			Key:   key,
			Value: bp,
			TTL:   0,
		})
	if err != nil {
		return bp, CacheError(errors.New("Failed to set post{" + key + "} to cache"))
	}
	return bp, nil
}

func (bc *BlogCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {

	var bpm []blog.BlogPostMeta
	err := bc.cache.Get(ctx, "PostsMeta", &bpm)

	if err == nil {
		return bpm, nil
	}
	if err != nil && err != cache.ErrCacheMiss {
		return nil, CacheError(fmt.Errorf("Failed to get postmeta: %w", err))

	}
	bpm, err = bc.persist.GetBlogPostsMeta(ctx)
	if err != nil {
		return nil, err
	}
	err = bc.cache.Set(
		&cache.Item{
			Ctx:   ctx,
			Key:   "PostsMeta",
			Value: bpm,
			TTL:   0,
		})
	if err != nil {
		return nil, CacheError(errors.New("Failed to add meta to cache"))
	}
	return bpm, nil
}

func (bc *BlogCache) Invalidate(ctx context.Context, id string) error {

	key := "post:" + id
	err := bc.cache.Delete(ctx, key)
	log.Print("Cache invalidates ", key)
	if err != nil {
		return CacheError(err)
	}
	err = bc.cache.Delete(ctx, "PostsMeta")

	if err != nil {
		return CacheError(err)
	}

	return nil
}
