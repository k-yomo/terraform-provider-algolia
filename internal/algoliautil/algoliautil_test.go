package algoliautil

import (
	"errors"
	"net/http"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
)

func TestIsRetryableError(t *testing.T) {
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
			name: "returns true if no more host error",
			args: args{
				err: errs.NewNoMoreHostToTryError(),
			},
			want: true,
		},
		{
			name: "returns false if not retryable error",
			args: args{err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.args.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
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
			if got := IsNotFoundError(tt.args.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}
