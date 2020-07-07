package stores

import (
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type BlogStore interface {
	GetBlogPosts() ([]BlogPost, error)
	GetBlogPost(string) (BlogPost, error)
}

type DropboxBlog struct {
	path string
	conf *dropbox.Config
}

type BlogPost struct {
	ID      string
	Title   string
	Content string
}

type BlogPosts []BlogPost

func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewDropboxBlogStore(key string) *DropboxBlog {

	return &DropboxBlog{path: "", conf: &dropbox.Config{Token: key, LogLevel: dropbox.LogOff}}
}

// GetBlogs retrieves files from my dropbox folder
func (dbx *DropboxBlog) GetBlogPosts() ([]BlogPost, error) {
	blogs := make([]BlogPost, 0)
	fileClient := files.New(*dbx.conf)

	arg := files.NewListFolderArg(dbx.path)

	res, err := fileClient.ListFolder(arg)
	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(fileClient, dbx.path)
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

			res, err = fileClient.ListFolderContinue(arg)
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
			blogs = append(blogs, BlogPost{ID: f.Id, Title: f.PathDisplay})
		case *files.FolderMetadata:
			fmt.Println("Folder:", i, f)
		default:
			fmt.Println("Default:", i, f)

		}

	}

	return blogs, nil

}

func (dbx *DropboxBlog) GetBlogPost(id string) (BlogPost, error) {
	return BlogPost{}, nil
}
