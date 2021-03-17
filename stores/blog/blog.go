package blog

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zaker/anachrome-be/stores/dropbox"
	"gopkg.in/yaml.v2"
)

type BlogStore interface {
	GetBlogPostsMeta(context.Context) ([]BlogPostMeta, error)
	GetBlogPost(context.Context, string) (BlogPost, error)
}

type DropboxBlog struct {
	client      *dropbox.Client
	UpdatesChan chan string
}

type BlogPostMeta struct {
	Title     string    `json:"title,omitempty"`
	ID        string    `json:"id,omitempty"`
	Published time.Time `json:"published,omitempty"`
	Updated   time.Time `json:"updated,omitempty"`
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

			content, _, err := dbx.client.GetFileContent(context.Background(), id)
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

			err = dbx.client.UpdateEntryProperties(context.Background(), ent, am)
			if err != nil {
				log.Println("updating anachrome meta", err)
				continue
			}
			dbx.UpdatesChan <- id
		}

	}
}

// NewDropboxBlogStore creates a dropbox blog store and initializes a syncing client
func NewDropboxBlogStore(client *http.Client, key, basePath, metadataID string) *DropboxBlog {

	c := dropbox.NewClient(client, key, basePath, metadataID)
	uc := make(chan string, 1)
	dbxBlog := &DropboxBlog{
		client:      c,
		UpdatesChan: uc}

	go dbxBlog.updateFileMetadata()
	return dbxBlog
}

// GetBlogPostsMeta lists files metadata
func (dbx *DropboxBlog) GetBlogPostsMeta(ctx context.Context) ([]BlogPostMeta, error) {
	meta := make([]BlogPostMeta, 0)
	folder, err := dbx.client.ListMainFolder(ctx)

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
			Updated:   ent.ClientModified,
		})
	}
	return meta, nil

}

func readAnachromeMetaFromContent(content []byte) (*dropbox.AnachromeMeta, int, error) {
	if len(content) < 8 {
		return nil, -1, fmt.Errorf("Content to short to include metadata")
	}
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
		Published: c.Published,
	}, idx + 7, nil
}

func (dbx *DropboxBlog) GetBlogPost(ctx context.Context, id string) (BlogPost, error) {

	content, filemeta, err := dbx.client.GetFileContent(ctx, id)
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
	blogPost.Meta.Updated = filemeta.ClientModified

	return blogPost, nil
}
