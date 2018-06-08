package spa

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"os"
)

//SPA I
type SPA struct {
	//PushFiles files to be pushed from index.html
	PushFiles []string
	//IndexHtml path to indexhtml
	IndexPath string

	appDir string
}

func getIntegrity(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return "sha256-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

//New initializes SPA
func New(appDir string) *SPA {

	tmpIndex := ".tmp/index.html"
	if _, err := os.Stat(".tmp"); os.IsNotExist(err) {
		os.Mkdir(".tmp", 0777)
	}

	return &SPA{[]string{}, tmpIndex, appDir}
}
