package provider

import (
	"testing"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/region"
)

func Test_parseImportRegionAndId(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}
	tests := []struct {
		name       string
		args       args
		wantRegion region.Region
		wantID     string
		wantErr    bool
	}{
		{
			name:       "parse region and id",
			args:       args{id: "eu/test"},
			wantRegion: region.EU,
			wantID:     "test",
		},
		{
			name:       "parse id",
			args:       args{id: "test"},
			wantRegion: "",
			wantID:     "test",
		},
		{
			name:    "invalid region",
			args:    args{id: "asia/test"},
			wantErr: true,
		},
		{
			name:    "invalid format",
			args:    args{id: "us/asia/test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegion, id, err := parseImportRegionAndId(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseImportRegionAndId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRegion != tt.wantRegion {
				t.Errorf("parseImportRegionAndId() gotRegion = %v, want %v", gotRegion, tt.wantRegion)
			}
			if id != tt.wantID {
				t.Errorf("parseImportRegionAndId() id = %v, want %v", id, tt.wantID)
			}
		})
	}
}
