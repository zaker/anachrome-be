package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
	"github.com/zaker/anachrome-be/stores/blog"
)

type RedisBlogCache struct {
	persist blog.BlogStore
	cache   *cache.Cache
}

func NewRedisBlogCache(p blog.BlogStore, redishost string) (*RedisBlogCache, error) {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": redishost,
		},
	})

	cache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &RedisBlogCache{persist: p, cache: cache}, nil
}

func (rbc *RedisBlogCache) GetBlogPost(ctx context.Context, id string) (blog.BlogPost, error) {

	var bp blog.BlogPost
	key := "post:" + id
	err := rbc.cache.Get(ctx, key, &bp)

	if err == nil {
		return bp, nil
	}
	if err != nil && err != cache.ErrCacheMiss {
		return bp, CacheError(errors.New("Failed to get post{" + key + "} "))
	}

	bp, err = rbc.persist.GetBlogPost(ctx, id)
	if err != nil {
		return bp, err
	}
	err = rbc.cache.Set(
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

func (rbc *RedisBlogCache) GetBlogPostsMeta(ctx context.Context) ([]blog.BlogPostMeta, error) {

	var bpm []blog.BlogPostMeta
	err := rbc.cache.Get(ctx, "PostsMeta", &bpm)

	if err == nil {
		return bpm, nil
	}
	if err != nil && err != cache.ErrCacheMiss {
		return nil, CacheError(fmt.Errorf("Failed to get postmeta: %w", err))

	}
	bpm, err = rbc.persist.GetBlogPostsMeta(ctx)
	if err != nil {
		return nil, err
	}
	err = rbc.cache.Set(
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

func (rbc *RedisBlogCache) Invalidate(ctx context.Context, id string) error {

	key := "post:" + id
	err := rbc.cache.Delete(ctx, key)
	log.Print("Cache invalidates ", key)
	if err != nil {
		return CacheError(err)
	}
	err = rbc.cache.Delete(ctx, "PostsMeta")

	if err != nil {
		return CacheError(err)
	}

	return nil
}
