package stores

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"gopkg.in/yaml.v2"
)

const propertyGroupTemplateID = "ANACHROME_BLOG"

type BlogStore interface {
	GetBlogPostsMeta() ([]BlogPostMeta, error)
	GetBlogPost(string) (BlogPost, error)
}

type DropboxBlog struct {
	path   string
	client files.Client
	ticker *time.Ticker
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

type BlogPosts []BlogPost

func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg("/" + path + ".md")

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (dbx *DropboxBlog) updateFileMetadata() {
	cursor := ""
	var err error
	for range dbx.ticker.C {
		_, cursor, err = dbx.getFilesMetadata(cursor)
		if err != nil {
			log.Println("update failed", err)
			continue
		}
	}
}

func NewDropboxBlogStore(key string) *DropboxBlog {

	client := files.New(dropbox.Config{Token: key, LogLevel: dropbox.LogOff})
	dbxBlog := &DropboxBlog{path: "", client: client, ticker: time.NewTicker(time.Second)}
	go dbxBlog.updateFileMetadata()
	return dbxBlog
}

func (dbx *DropboxBlog) getFilesMetadata(cursor string) ([]*files.FileMetadata, string, error) {

	var entries []files.IsMetadata
	var err error
	var res *files.ListFolderResult
	if cursor == "" {
		arg := files.NewListFolderArg(dbx.path)

		res, err = dbx.client.ListFolder(arg)

		cursor = res.Cursor
	} else {
		res = &files.ListFolderResult{Entries: make([]files.IsMetadata, 0), Cursor: cursor, HasMore: true}
	}

	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx.client, dbx.path)
				entries = []files.IsMetadata{metaRes}
			} else {
				return nil, cursor, err
			}
		default:
			return nil, cursor, err
		}

	} else {
		entries = res.Entries

		for res.HasMore {

			arg := files.NewListFolderContinueArg(cursor)

			res, err = dbx.client.ListFolderContinue(arg)
			cursor = res.Cursor
			if err != nil {
				return nil, cursor, err
			}

			entries = append(entries, res.Entries...)
		}
	}
	fmd := make([]*files.FileMetadata, 0)
	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			fmd = append(fmd, f)
		}

	}
	return fmd, cursor, nil
}

func convertFileMetadata(fm *files.FileMetadata) BlogPostMeta {
	bpm := BlogPostMeta{Path: trimDbxPath(fm.PathLower)}
	for _, pg := range fm.PropertyGroups {
		if pg.TemplateId == propertyGroupTemplateID {
			title := ""
			published := ""
			for _, field := range pg.Fields {
				if field.Name == "title" {
					title = field.Value
				}
				if field.Name == "published" {
					published = field.Value
				}

			}
			if title != "" && published != "" {
				bpm.Title = title
			}

		}

	}
	return bpm
}

// GetBlogPosts retrieves files from the main dropbox folder
func (dbx *DropboxBlog) GetBlogPostsMeta() ([]BlogPostMeta, error) {
	meta := make([]BlogPostMeta, 0)
	fmd, _, err := dbx.getFilesMetadata("")
	if err != nil {
		return nil, err
	}
	for _, f := range fmd {

		meta = append(meta, convertFileMetadata(f))
	}
	return meta, nil

}

func trimDbxPath(dbxPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(dbxPath, "/"), ".md")
}

func readContentMeta(content []byte) (*BlogPost, error) {
	contentString := string(content)
	if contentString[:4] != "---\n" {

		return nil, fmt.Errorf("Couldn't find metadata prelude")
	}

	idx := strings.Index(contentString[4:], "---")
	if idx == -1 {
		return nil, fmt.Errorf("Couldn't find metadata postlude")
	}
	var c ContentMeta

	data := contentString[4 : idx+4]
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {

		return nil, fmt.Errorf("cannot unmarshal data %w", err)
	}

	return &BlogPost{
		Content: strings.TrimSpace(contentString[idx+4+3+1:]),
		Meta: BlogPostMeta{
			Title:     c.Title,
			Published: &c.Published},
	}, nil
}

func (dbx *DropboxBlog) GetBlogPost(qPath string) (BlogPost, error) {
	path := "/" + qPath + ".md"
	blogPost := BlogPost{}
	arg := files.NewDownloadArg(path)

	filemeta, rc, err := dbx.client.Download(arg)
	defer rc.Close()
	if err != nil {
		switch e := err.(type) {
		case files.DownloadAPIError:

			return blogPost, e

		default:
			return blogPost, err
		}

	} else {
		fmt.Println(filemeta)
		content, err := ioutil.ReadAll(rc)

		if err != nil {
			return blogPost, err
		}
		blogPost.Meta.Path = trimDbxPath(filemeta.PathLower)

		tmpBP, err := readContentMeta(content)
		if err != nil {
			return blogPost, err
		}

		blogPost.Meta.Title = tmpBP.Meta.Title
		blogPost.Meta.Published = tmpBP.Meta.Published
		blogPost.Meta.Updated = &filemeta.ClientModified

	}

	return blogPost, nil
}
