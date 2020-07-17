package stores

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zaker/anachrome-be/stores/dropbox"
	"gopkg.in/yaml.v2"
)

type BlogStore interface {
	GetBlogPostsMeta() ([]BlogPostMeta, error)
	GetBlogPost(string) (BlogPost, error)
}

type DropboxBlog struct {
	client *dropbox.Client
}

type BlogPostMeta struct {
	Title     string     `json:"title,omitempty"`
	ID        string     `json:"id,omitempty"`
	Published *time.Time `json:"published,omitempty"`
	Updated   *time.Time `json:"updated,omitempty"`
}

type BlogPost struct {
	Meta    BlogPostMeta
	Content string
}

type ContentMeta struct {
	Title     string    `yaml:"title"`
	Published time.Time `yaml:"date"`
}

func (dbx *DropboxBlog) updateFileMetadata() {

	doneChan := make(chan struct{})
	entriesChan := make(chan dropbox.EntryMetadata)
	go func() {
		err := dbx.client.SubscribeMainFolder(entriesChan, doneChan)
		if err != nil {
			log.Println("subscribing to metadata failed", err)
		}
	}()

	for ent := range entriesChan {

		am, err := dbx.client.AnachromeMeta(ent)
		if err != nil {
			log.Println("getting anachrome meta failed", err)
			continue
		}
		if ent.ContentHash != am.Hash {
			id := dbx.client.GetID(ent)
			if err != nil {
				log.Println("getting file content", err)
				continue
			}
			content, _, err := dbx.client.GetFileContent(id)
			if err != nil {
				log.Println("getting file content", err)
				continue
			}
			meta, _, err := readAnachromeMetaFromContent(content)
			if err != nil {
				log.Println("reading anachrome meta", err)
				continue
			}

			am.Title = meta.Title
			am.Published = meta.Published
			am.Hash = ent.ContentHash

			err = dbx.client.UpdateEntryProperties(ent, am)
			if err != nil {
				log.Println("updating anachrome meta", err)
				continue
			}
		}

	}
}

// NewDropboxBlogStore creates a dropbox blog store and initializes a syncing client
func NewDropboxBlogStore(key, basePath, metadataID string) *DropboxBlog {

	client := dropbox.NewClient(key, basePath, metadataID)

	dbxBlog := &DropboxBlog{
		client: client}
	go dbxBlog.updateFileMetadata()
	return dbxBlog
}

// GetBlogPostsMeta lists files metadata
func (dbx *DropboxBlog) GetBlogPostsMeta() ([]BlogPostMeta, error) {
	meta := make([]BlogPostMeta, 0)
	folder, err := dbx.client.ListMainFolder()

	if err != nil {
		return nil, err
	}
	for _, ent := range folder.Entries {

		am, err := dbx.client.AnachromeMeta(ent)
		if err != nil {
			return nil, err
		}
		if am.Published.Sub(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() < 0 {
			continue
		}

		meta = append(meta, BlogPostMeta{
			Title:     am.Title,
			Published: am.Published,
			ID:        dbx.client.GetID(ent),
			Updated:   &ent.ClientModified,
		})
	}
	return meta, nil

}

func readAnachromeMetaFromContent(content []byte) (*dropbox.AnachromeMeta, int, error) {
	contentString := string(content)
	if contentString[:4] != "---\n" {

		return nil, -1, fmt.Errorf("Couldn't find metadata prelude")
	}

	idx := strings.Index(contentString[4:], "---")
	if idx == -1 {
		return nil, -1, fmt.Errorf("Couldn't find metadata postlude")
	}
	var c ContentMeta

	data := contentString[4 : idx+4]
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {

		return nil, idx + 4, fmt.Errorf("cannot unmarshal data %w", err)
	}
	return &dropbox.AnachromeMeta{
		Title:     c.Title,
		Published: &c.Published,
	}, idx + 7, nil
}

func (dbx *DropboxBlog) GetBlogPost(id string) (BlogPost, error) {

	content, filemeta, err := dbx.client.GetFileContent(id)
	blogPost := BlogPost{}
	if err != nil {
		return blogPost, err

	}

	fmt.Println(filemeta)

	if err != nil {
		return blogPost, err
	}
	blogPost.Meta.ID = dbx.client.GetID(*filemeta)

	meta, contentStart, err := readAnachromeMetaFromContent(content)
	if err != nil {
		return blogPost, err
	}
	blogPost.Content = strings.TrimSpace(string(content[contentStart:]))
	blogPost.Meta.Title = meta.Title
	blogPost.Meta.Published = meta.Published
	blogPost.Meta.Updated = &filemeta.ClientModified

	return blogPost, nil
}
