package algoliautil

import (
	"errors"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
	"net/http"
	"testing"
)

func TestIsAlgoliaNotFoundError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if not found error",
			args: args{
				err: errs.AlgoliaErr{
					Message: "not found",
					Status:  http.StatusNotFound,
				},
			},
			want: true,
		},
		{
			name: "returns false if not found error",
			args: args{
				err: errs.AlgoliaErr{
					Message: "bad request",
					Status:  http.StatusBadRequest,
				},
			},
			want: false,
		},
		{
			name: "returns false if not algolia error",
			args: args{err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAlgoliaNotFoundError(tt.args.err); got != tt.want {
				t.Errorf("IsAlgoliaNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}
