package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/zaker/anachrome-be/stores/blog"
)

type BlogPostMeta struct {
	blog.BlogPostMeta
	Path string `json:"path"`
}

func WantsHTML(header http.Header) bool {

	for _, v := range header.Values("Accept") {
		headers := strings.Split(v, " ")
		for _, h := range headers {
			if h == "text/html" {
				return true
			}
		}
	}
	return false
}

func BlogsToHTML(blogs []BlogPostMeta) (string, error) {
	sb := &bytes.Buffer{}
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Blog posts</title>
	</head>
	<body>
		{{range .}}<a href="{{ .Path }}">{{ .Title }}</a><br>{{else}}<div><strong>no blogs</strong></div>{{end}}
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parsing html template: %w", err)
	}
	err = t.Execute(sb, blogs)
	if err != nil {
		return "", fmt.Errorf("executing html template: %w", err)
	}
	htmlBytes, err := io.ReadAll(sb)
	if err != nil {
		return "", fmt.Errorf("read html template: %w", err)
	}
	return string(htmlBytes), nil
}

func BlogToHTML(blogs blog.BlogPost) (string, error) {
	sb := &bytes.Buffer{}
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{ .Meta.Title }}</title>
	</head>
	<body>
		<h1>{{ .Meta.Title }}</h1>
		<h2>Published: {{ .Meta.Published }}</h2>
		<p> {{ .Content }}</p>
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parsing html template: %w", err)
	}
	err = t.Execute(sb, blogs)
	if err != nil {
		return "", fmt.Errorf("executing html template: %w", err)
	}
	htmlBytes, err := io.ReadAll(sb)
	if err != nil {
		return "", fmt.Errorf("read html template: %w", err)
	}
	return string(htmlBytes), nil
}
