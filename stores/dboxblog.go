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

type BlogStore interface {
	GetBlogPosts() ([]BlogPost, error)
	GetBlogPost(string) (BlogPost, error)
}

type DropboxBlog struct {
	path   string
	client files.Client
}

type BlogPost struct {
	Path      string
	Title     string
	Content   string
	Published time.Time
	Updated   time.Time
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

func NewDropboxBlogStore(key string) *DropboxBlog {

	client := files.New(dropbox.Config{Token: key, LogLevel: dropbox.LogOff})
	return &DropboxBlog{path: "", client: client}
}

// GetBlogs retrieves files from my dropbox folder
func (dbx *DropboxBlog) GetBlogPosts() ([]BlogPost, error) {
	blogs := make([]BlogPost, 0)

	arg := files.NewListFolderArg(dbx.path)

	res, err := dbx.client.ListFolder(arg)
	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx.client, dbx.path)
				entries = []files.IsMetadata{metaRes}
			} else {
				return blogs, err
			}
		default:
			return blogs, err
		}

	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.client.ListFolderContinue(arg)
			if err != nil {
				return blogs, err
			}

			entries = append(entries, res.Entries...)
		}
	}
	for i, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			fmt.Println("File:", i, f)
			blogs = append(blogs, BlogPost{Path: trimDbxPath(f.PathLower), Title: f.Name})
		case *files.FolderMetadata:
			fmt.Println("Folder:", i, f)
		default:
			fmt.Println("Default:", i, f)

		}

	}

	return blogs, nil

}

func trimDbxPath(dbxPath string) string {
	return strings.TrimSuffix(strings.TrimPrefix(dbxPath, "/"), ".md")
}

func setContentMeta(blogPost *BlogPost, content []byte) {
	contentString := string(content)
	if contentString[:4] != "---\n" {
		log.Printf("no metadata found\n")
		return
	}

	idx := strings.Index(contentString[4:], "---")
	if idx == -1 {
		log.Printf("no metadata found\n")
		return
	}
	var c ContentMeta

	data := contentString[4 : idx+4]
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Printf("cannot unmarshal data: %v\n", err)
		return
	}
	blogPost.Title = c.Title
	blogPost.Content = strings.TrimSpace(contentString[idx+4+3+1:])
	blogPost.Published = c.Published

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
		blogPost.Path = trimDbxPath(filemeta.PathLower)

		setContentMeta(&blogPost, content)
		blogPost.Updated = filemeta.ClientModified

	}

	return blogPost, nil
}
