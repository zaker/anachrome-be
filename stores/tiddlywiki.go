package stores

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type TiddlerFileStore struct {
	BasePath string
}

type Tiddler struct {
	Rev  int
	Meta string
	Text string
}

func (ts *TiddlerFileStore) List(w io.Writer) error {

	var buf bytes.Buffer
	sep := ""
	buf.WriteString("[")
	for {
		var t Tiddler

		if len(t.Meta) == 0 {
			break
		}
		meta := t.Meta

		if strings.Contains(meta, `"$:/tags/Macro"`) {
			var js map[string]interface{}
			err := json.Unmarshal([]byte(meta), &js)
			if err != nil {
				continue
			}
			js["text"] = string(t.Text)
			data, err := json.Marshal(js)
			if err != nil {
				continue
			}
			meta = string(data)
		}

		buf.WriteString(sep)
		// sep = ","
		buf.WriteString(meta)
		break
	}
	buf.WriteString("]")
	_, err := w.Write(buf.Bytes())
	return err
}

func (ts *TiddlerFileStore) Store(r io.Reader, id string) (string, error) {
	k := GenKey(id)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(ts.BasePath+"/tiddlers/"+k, data, 0644)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("\"bag/%s/%d:%x\"", id, 1, md5.Sum(data)), nil
}

func (ts *TiddlerFileStore) Load(w io.Writer, id string) error {
	k := GenKey(id)
	rc, err := os.Open(ts.BasePath + "/tiddlers/" + k)
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(w, rc)
	return err
}

func (ts *TiddlerFileStore) Index(w io.Writer) error {

	rc, err := os.Open(ts.BasePath + "/index.html")
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(w, rc)
	return err
}

func (ts *TiddlerFileStore) Status(w io.Writer) error {

	_, err := w.Write([]byte(`{"username": "zaker", "space": {"recipe": "all"}}`))

	return err

}

func (ts *TiddlerFileStore) Delete() error {

	return nil

}

func GenKey(id string) string {

	return fmt.Sprintf("%x.tid.json", md5.Sum([]byte(id)))
}
