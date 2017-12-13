package index

import (
	"html/template"
	"log"
	"os"
	"strings"
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
	var (
		funcs     = template.FuncMap{"join": strings.Join}
		guardians = []string{"Gamora", "Groot", "Nebula", "Rocket", "Star-Lord"}
	)

	masterTmpl, err := template.New("index").Funcs(funcs).Parse(indexTmplt)
	if err != nil {
		log.Fatal(err)
	}
	if err := masterTmpl.Execute(os.Stdout, guardians); err != nil {
		log.Fatal(err)
	}
	return ""
}
