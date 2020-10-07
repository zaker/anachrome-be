package controllers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zaker/anachrome-be/stores"
)

type Blog struct {
	blogs    stores.BlogStore
	basePath string
}

type BlogPostMeta struct {
	stores.BlogPostMeta
	Path string `json:"path"`
}

func NewBlog(blogs stores.BlogStore, basePath string) *Blog {
	return &Blog{blogs, basePath}
}

func (b *Blog) ListBlogPosts(c echo.Context) error {

	blogPosts := make([]BlogPostMeta, 0)
	bpm, err := b.blogs.GetBlogPostsMeta(context.TODO())
	if err != nil {
		return err
	}

	for _, bpmEnt := range bpm {
		bm := BlogPostMeta{BlogPostMeta: bpmEnt,
			Path: b.basePath + "/blog/" + bpmEnt.ID}
		blogPosts = append(blogPosts, bm)
	}

	return c.JSON(http.StatusOK, blogPosts)
}

func (b *Blog) GetBlogPost(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return c.JSON(http.StatusNotFound, id)
	}
	post, err := b.blogs.GetBlogPost(context.TODO(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, post)
}
