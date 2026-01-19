package controllers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/zaker/anachrome-be/services"
	"github.com/zaker/anachrome-be/stores/blog"
)

type Blog struct {
	blogs    blog.BlogStore
	basePath string
}

func NewBlog(blogs blog.BlogStore, basePath string) *Blog {
	return &Blog{blogs, basePath}
}

func (b *Blog) ListBlogPosts(c *echo.Context) error {

	blogPosts := make([]services.BlogPostMeta, 0)
	bpm, err := b.blogs.GetBlogPostsMeta(context.TODO())
	if err != nil {
		return err
	}

	for _, bpmEnt := range bpm {
		bm := services.BlogPostMeta{BlogPostMeta: bpmEnt,
			Path: b.basePath + "/blog/" + bpmEnt.ID}
		blogPosts = append(blogPosts, bm)
	}

	if services.WantsHTML(c.Request().Header) {
		htmlStr, err := services.BlogsToHTML(blogPosts)
		if err != nil {
			return err
		}
		return c.HTML(http.StatusOK, htmlStr)
	}
	return c.JSON(http.StatusOK, blogPosts)
}

func (b *Blog) GetBlogPost(c *echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return c.JSON(http.StatusNotFound, id)
	}
	post, err := b.blogs.GetBlogPost(context.TODO(), id)
	if err != nil {
		return err
	}

	if services.WantsHTML(c.Request().Header) {
		htmlStr, err := services.BlogToHTML(post)
		if err != nil {
			return err
		}
		return c.HTML(http.StatusOK, htmlStr)
	}
	return c.JSON(http.StatusOK, post)
}
