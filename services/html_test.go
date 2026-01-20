package services

import (
	"net/http"
	"testing"
	"time"

	"github.com/zaker/anachrome-be/stores/blog"
)

func TestWantsHTML(t *testing.T) {

	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{"No header", http.Header{}, false},
		{"Accepted", http.Header{"Accept": []string{"text/html application/octet-stream"}}, true},
		{"Chrome wants html", http.Header{"Accept": []string{"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"}}, true},
		{"Doubly Accepted", http.Header{"Accept": []string{"application/octet-stream", "text/html"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WantsHTML(tt.header); got != tt.want {
				t.Errorf("WantsHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlogsToHTML(t *testing.T) {

	tests := []struct {
		name    string
		blogs   []BlogPostMeta
		want    string
		wantErr bool
	}{
		{"Empty list", []BlogPostMeta{},
			`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Blog posts</title>
	</head>
	<body>
		<div><strong>no blogs</strong></div>
	</body>
</html>`,
			false},
		{"List 2", []BlogPostMeta{
			{
				blog.BlogPostMeta{Title: "Foo", Published: time.Date(2021, 3, 18, 10, 27, 0, 0, time.UTC)},
				"http://example.com/blog/Foo",
			},
			{
				blog.BlogPostMeta{Title: "Bar", Published: time.Date(2021, 2, 18, 10, 27, 0, 0, time.UTC)},
				"http://example.com/blog/Bar",
			},
		},
			`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Blog posts</title>
	</head>
	<body>
		<a href="http://example.com/blog/Foo">Foo</a><br><a href="http://example.com/blog/Bar">Bar</a><br>
	</body>
</html>`,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BlogsToHTML(tt.blogs)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlogsToHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BlogsToHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlogToHTML(t *testing.T) {

	tests := []struct {
		name    string
		blog    blog.BlogPost
		want    string
		wantErr bool
	}{
		{
			"Should return blog html",
			blog.BlogPost{
				Meta:    blog.BlogPostMeta{Title: "Foo", Published: time.Date(2021, 2, 18, 10, 27, 0, 0, time.UTC)},
				Content: "Foo text",
			},
			`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Foo</title>
	</head>
	<body>
		<h1>Foo</h1>
		<h2>Published: 2021-02-18 10:27:00 +0000 UTC</h2>
		<p> Foo text</p>
	</body>
</html>`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BlogToHTML(tt.blog)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlogToHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BlogToHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
