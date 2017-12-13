package index

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/html"
)

var indexTmplt = `<!doctype html>
<html lang="en"><head>
    <meta charset="utf-8">
    <title>Anachrome</title>
    <base href="/">
    <meta name="viewport" content="width=device-width,initial-scale=1">
    <link rel="icon" type="image/x-icon" href="favicon.ico">
    <link href="styles.d41d8cd98f00b204e980.bundle.css" rel="stylesheet" />
</head><body>
    <app-root></app-root>
    <script type="text/javascript" src="inline.43bdfccbf94fc813c9b1.bundle.js"></script>
    <script type="text/javascript" src="polyfills.43a6a16e791d2caa0484.bundle.js"></script>
    
</body></html>`

var scriptTag = `<script type="text/javascript" src="main.acdbc32b55ccec0d850f.bundle.js"></script>`
var stuleTag = `<link href="styles.d41d8cd98f00b204e980.bundle.css" rel="stylesheet" />`

//HTML creates an index.html from a set of angular app files and adds security headers
func HTML(rootDir string) string {
	s := `<p>Links:</p><ul><li><a href="foo">Foo</a><li><a href="/bar/baz">BarBaz</a></ul>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		log.Fatal(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "link") {
			for _, a := range n.Attr {
				if n.Data == "script" && a.Key == "src" {
					fmt.Println("hashing:", a.Val)
					break
				}

				if n.Data == "link" && a.Key == "href" {
					fmt.Println("hashing:", a.Val)
					break
				}
			}
			n.Attr = append(n.Attr, html.Attribute{Key: "integrity"})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return ""
}
