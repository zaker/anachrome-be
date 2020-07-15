package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zaker/anachrome-be/stores"
)

type Blog struct {
	blogs    stores.BlogStore
	basePath string
}

func NewBlog(blogs stores.BlogStore, basePath string) *Blog {
	return &Blog{blogs, basePath}
}

func (b *Blog) ListBlogPosts(c echo.Context) error {

	bpm, err := b.blogs.GetBlogPostsMeta()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, bpm)
}

func (b *Blog) GetBlogPost(c echo.Context) error {
	path := c.Param("path")
	if len(path) == 0 {
		return c.JSON(http.StatusNotFound, path)
	}
	post, err := b.blogs.GetBlogPost(path)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, post)
}
