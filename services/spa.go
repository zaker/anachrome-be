package services

import (
	"log"
	"os"

	"golang.org/x/net/html"
)

//SPA I
type SPA struct {
	//PushFiles files to be pushed from index.html
	PushFiles []string
	//IndexHtml path to indexhtml
	IndexPath string

	appDir string
}

//New initializes SPA
func NewSPA(appDir string) (*SPA, error) {

	tmpIndex := ".tmp/index.html"
	if _, err := os.Stat(".tmp"); os.IsNotExist(err) {
		err := os.Mkdir(".tmp", 0777)
		if err != nil {
			return nil, err
		}
	}

	return &SPA{[]string{}, tmpIndex, appDir}, nil
}

func (s *SPA) addPushFiles(n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "link" || n.Data == "base") {
		src := ""
		base := ""
		for _, a := range n.Attr {
			if n.Data == "script" && a.Key == "src" {
				src = a.Val
				break
			}

			if n.Data == "link" && a.Key == "href" {
				src = a.Val
				break
			}
			if n.Data == "base" && a.Key == "href" {
				base = a.Val
				break
			}
		}
		if len(src) > 0 {
			s.PushFiles = append(s.PushFiles, base+src)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.addPushFiles(c)
	}
}

//IndexParse creates an index.html from a set of angular app files and adds security headers
func (s *SPA) IndexParse() {
	idx, err := os.Open(s.appDir + "index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Close()
	tmpIdx, err := os.Create(s.IndexPath)
	if err != nil {
		log.Fatal(err)
	}
	defer tmpIdx.Close()

	doc, err := html.Parse(idx)
	if err != nil {
		log.Fatal(err)
	}

	s.addPushFiles(doc)
	err = html.Render(tmpIdx, doc)
	if err != nil {
		log.Fatal(err)
	}
}
