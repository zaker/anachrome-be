package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type FolderMetadata struct {
	Entries []EntryMetadata `json:"entries"`
	Cursor  string          `json:"cursor"`
	HasMore bool            `json:"has_more"`
}

type propertyGroup struct {
	TemplateID string  `json:"template_id"`
	Fields     []field `json:"fields"`
}

type propertyGroupUpdate struct {
	TemplateID        string  `json:"template_id"`
	AddOrUpdateFields []field `json:"add_or_update_fields"`
}

type field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type EntryMetadata struct {
	Tag            string           `json:".tag"`
	Name           string           `json:"name"`
	PathLower      string           `json:"path_lower"`
	PathDisplay    string           `json:"path_display"`
	ID             string           `json:"id"`
	ClientModified time.Time        `json:"client_modified"`
	ServerModified time.Time        `json:"server_modified"`
	Rev            string           `json:"rev"`
	Size           int              `json:"size"`
	IsDownloadable bool             `json:"is_downloadable"`
	PropertyGroups *[]propertyGroup `json:"property_groups,omitempty"`
	ContentHash    string           `json:"content_hash"`
}
type AnachromeMeta struct {
	Title     string
	Published *time.Time
	Hash      string
}

type Client struct {
	key                string
	basePath           string
	metadataTemplateID string
}

func NewClient(key, path, metadataTemplateID string) *Client {
	if path == "" {
		path = "/blog"
	}
	return &Client{key: key, basePath: path, metadataTemplateID: metadataTemplateID}
}

type listFolderArg struct {
	Path                        string               `json:"path"`
	IncludePropertyGroups       propertyGroupsFilter `json:"include_property_groups"`
	IncludeNonDownloadableFiles bool                 `json:"include_non_downloadable_files"`
	IncludeDeleted              bool                 `json:"include_deleted"`
}

type propertyGroupsFilter struct {
	Tag  string   `json:".tag"`
	List []string `json:"filter_some"`
}

func (c *Client) createFolderMetadataRequest() (*http.Request, error) {
	arg := &listFolderArg{
		Path: c.basePath,
		IncludePropertyGroups: propertyGroupsFilter{
			Tag:  "filter_some",
			List: []string{c.metadataTemplateID},
		},
		IncludeNonDownloadableFiles: false,
		IncludeDeleted:              false,
	}

	b, err := json.Marshal(arg)
	if err != nil {
		return nil, fmt.Errorf("Marshalling args: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://api.dropboxapi.com/2/files/list_folder",
		bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("Creating folder request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.key)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (c *Client) ListMainFolder(ctx context.Context) (*FolderMetadata, error) {
	client := &http.Client{}

	req, err := c.createFolderMetadataRequest()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("Requesting folder metadata: %w", err)
	}
	if resp.StatusCode != 200 {
		code, err := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Requesting folder metadata: %s , %w", code, err)
	}

	decoder := json.NewDecoder(resp.Body)
	data := &FolderMetadata{}
	err = decoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("Decoding response: %w", err)
	}
	return data, nil
}

func (c *Client) continueMainFolder(ctx context.Context, cursor string) (*FolderMetadata, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.dropboxapi.com/2/files/list_folder/continue",
		bytes.NewReader([]byte(fmt.Sprintf(`
		{
			"cursor":"%s"
		}`, cursor))))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Requesting folder metadata: %w", err)
	}
	if resp.StatusCode != 200 {
		code, err := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Requesting folder metadata: %s , %w", code, err)
	}

	decoder := json.NewDecoder(resp.Body)
	data := &FolderMetadata{}
	err = decoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("Decoding response: %w", err)
	}
	return data, nil
}

func (c *Client) SubscribeMainFolder(entriesChan chan<- EntryMetadata, done <-chan struct{}) error {

	initResults, err := c.ListMainFolder(context.Background())
	if err != nil {
		return err
	}
	for _, ent := range initResults.Entries {
		entriesChan <- ent
	}
	ticker := time.NewTicker(1 * time.Second)
	cursor := initResults.Cursor
	for range ticker.C {
		res, err := c.continueMainFolder(context.Background(), cursor)
		if err != nil {
			panic(err)
		}
		for _, ent := range res.Entries {
			entriesChan <- ent
		}
	}

	return nil
}

func (c *Client) AnachromeMeta(emd EntryMetadata) (AnachromeMeta, error) {

	am := AnachromeMeta{}
	if emd.PropertyGroups == nil {
		return am, nil
	}
	for _, pg := range *emd.PropertyGroups {
		if pg.TemplateID == c.metadataTemplateID {
			for _, field := range pg.Fields {
				if field.Name == "title" {
					am.Title = field.Value
				}
				if field.Name == "published" {
					d, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", field.Value)
					if err != nil {
						return am, fmt.Errorf("Published is not a date: %w", err)
					}
					am.Published = &d
				}
				if field.Name == "hash" {
					am.Hash = field.Value
				}

			}
			return am, nil

		}

	}
	return am, nil
}

type propertyArg struct {
	Path                 string                 `json:"path"`
	PropertyGroups       *[]propertyGroup       `json:"property_groups,omitempty"`
	UpdatePropertyGroups *[]propertyGroupUpdate `json:"update_property_groups,omitempty"`
}

func (c *Client) UpdateEntryProperties(ctx context.Context, ent EntryMetadata, am AnachromeMeta) error {
	client := &http.Client{}
	mode := "add"

	arg := propertyArg{
		Path: ent.PathLower,
	}

	if len(*ent.PropertyGroups) == 1 {
		mode = "update"
		arg.UpdatePropertyGroups = &[]propertyGroupUpdate{{
			TemplateID: c.metadataTemplateID,
			AddOrUpdateFields: []field{
				{Name: "title", Value: am.Title},
				{Name: "published", Value: am.Published.String()},
				{Name: "hash", Value: am.Hash},
			}}}
	} else {
		arg.PropertyGroups = &[]propertyGroup{{
			TemplateID: c.metadataTemplateID,
			Fields: []field{
				{Name: "title", Value: am.Title},
				{Name: "published", Value: am.Published.String()},
				{Name: "hash", Value: am.Hash},
			}}}
	}

	b, err := json.Marshal(arg)
	if err != nil {
		return fmt.Errorf("Marshalling args: %w", err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.dropboxapi.com/2/file_properties/properties/"+mode,
		bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("Creating file properties request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.key)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Add/Update file metadata: %w", err)
	}
	if resp.StatusCode != 200 {
		code, err := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Add/Update file metadata: %s , %w", code, err)
	}
	return nil
}

func (c *Client) GetFileContent(ctx context.Context, id string) ([]byte, *EntryMetadata, error) {

	path := c.basePath + "/" + id + ".md"
	client := &http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://content.dropboxapi.com/2/files/download", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Creating file content request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+c.key)
	req.Header.Add("Dropbox-API-Arg", fmt.Sprintf("{\"path\":\"%s\"}", path))
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Downloading file content: %w", err)
	}
	if resp.StatusCode != 200 {
		code, err := ioutil.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("Downloading file content: %s , %w", code, err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Reading file content: %w", err)
	}
	var meta EntryMetadata
	err = json.Unmarshal([]byte(resp.Header["Dropbox-Api-Result"][0]), &meta)
	if err != nil {
		return nil, nil, fmt.Errorf("Reading file content: %w", err)
	}
	return content, &meta, nil
}

func (c *Client) GetID(entry EntryMetadata) string {
	id := strings.TrimPrefix(entry.PathLower, c.basePath+"/")
	id = strings.TrimSuffix(id, ".md")
	return id
}
