package blog

import (
	"reflect"
	"testing"
	"time"

	"github.com/zaker/anachrome-be/stores/dropbox"
)

func Test_readAnachromeMetaFromContent(t *testing.T) {

	tests := []struct {
		name    string
		content []byte
		want    *dropbox.AnachromeMeta
		want1   int
		wantErr bool
	}{
		{"Empty file should retur error", []byte(""), nil, -1, true},
		{"CRLF should give error", []byte("---\r\npublished: 2006-01-02 15:04:05 +0000 UTC\r\ntitle: test\r\n---\r\n"), nil, -1, true},
		{
			"Should return meta",
			[]byte("---\ndate: 2021-03-17\ntitle: Test\n---\n"),
			&dropbox.AnachromeMeta{
				Title:     "Test",
				Published: time.Date(2021, 3, 17, 0, 0, 0, 0, time.UTC),
			},
			36,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := readAnachromeMetaFromContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("readAnachromeMetaFromContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readAnachromeMetaFromContent() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("readAnachromeMetaFromContent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
