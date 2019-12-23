package stores

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type TiddlerFileStore struct {
	BasePath string
}

type Tiddler struct {
	Rev  int    `json:"rev"`
	Meta string `json:"meta"`
	Text string `json:"text"`
}

func TiddlerFromFile(t *Tiddler, file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), &t)

}

func (ts *TiddlerFileStore) List(w io.Writer) error {

	files, err := filepath.Glob(path.Join(
		ts.BasePath,
		"tiddlers",
		fmt.Sprintf("*.tid.json")))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	sep := ""
	buf.WriteString("[")
	for _, f := range files {
		var t Tiddler
		err = TiddlerFromFile(&t, f)
		if err != nil {
			continue
		}
		if len(t.Meta) == 0 {
			continue
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
		sep = ","
		buf.WriteString(meta)

	}
	buf.WriteString("]")
	_, err = w.Write(buf.Bytes())
	return err
}

func (ts *TiddlerFileStore) Store(r io.Reader, id string) (string, error) {
	tp := ts.tiddlerPath(id)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if err != nil {
		return "", err
	}

	js["bag"] = "bag"

	rev := 1
	var old Tiddler
	if err := TiddlerFromFile(&old, tp); err == nil {
		rev = old.Rev + 1
	}
	js["revision"] = rev

	var t Tiddler
	text, ok := js["text"].(string)
	if ok {
		t.Text = text
	}
	delete(js, "text")
	t.Rev = rev
	meta, err := json.Marshal(js)
	if err != nil {

		return "", err
	}
	t.Meta = string(meta)

	tidContent, err := json.Marshal(t)
	if err != nil {

		return "", err
	}
	err = ioutil.WriteFile(tp, tidContent, 0644)
	if err != nil {
		return "", err
	}

	hp := ts.historyPath(id, t.Rev)
	err = ioutil.WriteFile(hp, tidContent, 0644)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("\"bag/%s/%d:%x\"", id, t.Rev, md5.Sum(data)), nil
}

func (ts *TiddlerFileStore) Load(w io.Writer, id string) error {

	var t Tiddler
	err := TiddlerFromFile(&t, ts.tiddlerPath(id))
	if err != nil {
		return err
	}
	var js map[string]interface{}
	err = json.Unmarshal([]byte(t.Meta), &js)
	if err != nil {
		return err
	}
	js["text"] = string(t.Text)
	data, err := json.Marshal(js)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

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

func (ts *TiddlerFileStore) Delete(id string) error {

	tp := ts.tiddlerPath(id)
	rc, err := os.Open(tp)
	if err != nil {
		return err
	}
	defer rc.Close()
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	var t Tiddler
	err = json.Unmarshal(data, &t)
	if err != nil {
		return err
	}
	t.Rev++
	t.Meta = ""
	t.Text = ""

	data, err = json.Marshal(t)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tp, data, 0644)
	if err != nil {
		return err
	}
	hp := ts.historyPath(id, t.Rev)
	err = ioutil.WriteFile(hp, data, 0644)
	if err != nil {
		return err
	}
	return nil

}

func GenKey(id string) string {

	return fmt.Sprintf("%x.tid.json", md5.Sum([]byte(id)))
}

func (ts *TiddlerFileStore) tiddlerPath(id string) string {
	return path.Join(
		ts.BasePath,
		"tiddlers",
		fmt.Sprintf("%x.tid.json",
			md5.Sum([]byte(id))))
}

func (ts *TiddlerFileStore) historyPath(id string, rev int) string {
	return path.Join(
		ts.BasePath,
		"history",
		fmt.Sprintf("%x_%d.tid.json",
			md5.Sum([]byte(id)), rev))
}
