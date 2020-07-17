package stores

import (
	"fmt"
	"log"
	"strings"
	"time"

	mydbx "github.com/zaker/anachrome-be/stores/dropbox"
	"gopkg.in/yaml.v2"
)

type BlogStore interface {
	GetBlogPostsMeta() ([]BlogPostMeta, error)
	GetBlogPost(string) (BlogPost, error)
}

type DropboxBlog struct {
	path        string
	myDbxClient *mydbx.Client
	ticker      *time.Ticker
}

type BlogPostMeta struct {
	Title     string     `json:"title,omitempty"`
	Path      string     `json:"path,omitempty"`
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
	entriesChan := make(chan mydbx.EntryMetadata)
	go func() {
		err := dbx.myDbxClient.SubscribeMainFolder(entriesChan, doneChan)
		if err != nil {
			log.Println("subscribing to metadata failed", err)
		}
	}()

	for ent := range entriesChan {

		am, err := dbx.myDbxClient.AnachromeMeta(ent)
		if err != nil {
			log.Println("getting anachrome meta failed", err)
			continue
		}
		if ent.ContentHash != am.Hash {
			log.Println("Needs update", ent)
			content, _, err := dbx.myDbxClient.GetFileContent(ent.PathLower)
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

			err = dbx.myDbxClient.UpdateEntryProperties(ent, am)
			if err != nil {
				log.Println("updating anachrome meta", err)
				continue
			}
		}

	}
}

// NewDropboxBlogStore creates a dropbox blog store and initializes a syncing client
func NewDropboxBlogStore(key, basePath, metadataID string) *DropboxBlog {

	myDbxClient := mydbx.NewClient(key, basePath, metadataID)

	dbxBlog := &DropboxBlog{
		path:        "",
		myDbxClient: myDbxClient,
		ticker:      time.NewTicker(time.Second)}
	go dbxBlog.updateFileMetadata()
	return dbxBlog
}

// GetBlogPostsMeta lists files metadata
func (dbx *DropboxBlog) GetBlogPostsMeta() ([]BlogPostMeta, error) {
	meta := make([]BlogPostMeta, 0)
	folder, err := dbx.myDbxClient.ListMainFolder()

	if err != nil {
		return nil, err
	}
	for _, ent := range folder.Entries {

		am, err := dbx.myDbxClient.AnachromeMeta(ent)
		if err != nil {
			return nil, err
		}

		meta = append(meta, BlogPostMeta{
			Title:     am.Title,
			Published: am.Published,
			Path:      ent.PathLower,
			Updated:   &ent.ClientModified,
		})
	}
	return meta, nil

}

func trimDbxPath(dbxPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(dbxPath, "/"), ".md")
}

func readAnachromeMetaFromContent(content []byte) (*mydbx.AnachromeMeta, int, error) {
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
	return &mydbx.AnachromeMeta{
		Title:     c.Title,
		Published: &c.Published,
	}, idx + 7, nil
}

func (dbx *DropboxBlog) GetBlogPost(qPath string) (BlogPost, error) {

	content, filemeta, err := dbx.myDbxClient.GetFileContent(qPath)
	blogPost := BlogPost{}
	if err != nil {
		return blogPost, err

	}

	fmt.Println(filemeta)

	if err != nil {
		return blogPost, err
	}
	blogPost.Meta.Path = trimDbxPath(filemeta.PathLower)

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
