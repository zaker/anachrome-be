package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zaker/anachrome-be/mocks"

	"github.com/zaker/anachrome-be/stores/blog"
)

func TestCache(t *testing.T) {

	wantBlog := blog.BlogPost{Meta: blog.BlogPostMeta{Title: "Foo"}}
	mbs := &mocks.MockBlogStore{
		GetBlogPostFunc: func(ctx context.Context, id string) (blog.BlogPost, error) {
			if id == "1" {
				return wantBlog, nil
			}
			return blog.BlogPost{}, fmt.Errorf("Can't find blog for id(%s)", id)
		},
	}
	c, err := NewInMemoryCache(mbs)
	if err != nil {
		t.Errorf("Initializing cache failed: %w", err)
	}

	gotBlog, err := c.GetBlogPost(context.Background(), "1")

	if err != nil {
		t.Errorf("Failed fetching exisitng item: %w", err)
	}
	callsToGBP := len(mbs.GetBlogPostCalls())

	assert.Equal(t, wantBlog, gotBlog)
	assert.Equal(t, 1, callsToGBP)

	_, err = c.GetBlogPost(context.Background(), "2")

	if err == nil {
		t.Errorf("Should fail when fetching non existing items")
	}

	gotBlog, err = c.GetBlogPost(context.Background(), "1")

	if err != nil {
		t.Errorf("Failed fetching exisitng item: %w", err)
	}
	callsToGBP = len(mbs.GetBlogPostCalls())

	assert.Equal(t, wantBlog, gotBlog)
	assert.Equal(t, 2, callsToGBP)

	c.Invalidate(context.Background(), "1")

	gotBlog, err = c.GetBlogPost(context.Background(), "1")

	if err != nil {
		t.Errorf("Failed fetching exisitng item: %w", err)
	}
	callsToGBP = len(mbs.GetBlogPostCalls())

	assert.Equal(t, wantBlog, gotBlog)
	assert.Equal(t, 3, callsToGBP)

}
