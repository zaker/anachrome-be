package dropbox

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	type args struct {
		client             *http.Client
		key                string
		path               string
		metadataTemplateID string
	}

	httpClient := http.DefaultClient

	tests := []struct {
		name string
		args args
		want *Client
	}{
		{"Should return new client",
			args{
				client:             httpClient,
				key:                "key",
				path:               "path",
				metadataTemplateID: "id",
			},
			&Client{
				client:             httpClient,
				key:                "key",
				basePath:           "path",
				metadataTemplateID: "id",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(
				tt.args.client,
				tt.args.key,
				tt.args.path,
				tt.args.metadataTemplateID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_AnachromeMeta(t *testing.T) {

	tmpDate := time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC)

	tests := []struct {
		name    string
		c       *Client
		emd     EntryMetadata
		want    AnachromeMeta
		wantErr bool
	}{
		{
			"Should create empty metadata",
			NewClient(nil, "key", "path", "metadata"),
			EntryMetadata{},
			AnachromeMeta{},
			false,
		},
		{
			"Should create with metadata",
			NewClient(nil, "key", "path", "metadata"),
			EntryMetadata{
				PropertyGroups: &[]propertyGroup{
					{TemplateID: "metadata",
						Fields: []field{
							{Name: "title", Value: "title 1"},
							{Name: "published", Value: "2006-01-02 15:04:05 +0000 UTC"},
							{Name: "hash", Value: "value2"},
						}},
				},
			},
			AnachromeMeta{
				Title:     "title 1",
				Published: tmpDate,
				Hash:      "value2",
			},
			false,
		},
		{
			"Should fail on date metadata",
			NewClient(nil, "key", "path", "metadata"),
			EntryMetadata{
				PropertyGroups: &[]propertyGroup{
					{TemplateID: "metadata",
						Fields: []field{
							{Name: "title", Value: "title 1"},
							{Name: "published", Value: "2006-01-02 a:04:05"},
							{Name: "hash", Value: "value2"},
						}},
				},
			},
			AnachromeMeta{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.AnachromeMeta(tt.emd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AnachromeMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.AnachromeMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}
