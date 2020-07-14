package controllers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/zaker/anachrome-be/stores"
)

type Blog struct {
	blogs    stores.BlogStore
	basePath string
}

type BlogPostMeta struct {
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
}

func NewBlog(blogs stores.BlogStore, basePath string) *Blog {
	return &Blog{blogs, basePath}
}

func (b *Blog) ListBlogPosts(c echo.Context) error {

	blogPosts, err := b.blogs.GetBlogPosts()
	if err != nil {
		return err
	}

	bpm := make([]BlogPostMeta, 0)
	for _, p := range blogPosts {
		bpm = append(bpm, BlogPostMeta{
			Title: p.Title,
			Path:  fmt.Sprintf("%s/blog/%s", b.basePath, p.Path),
		})

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
