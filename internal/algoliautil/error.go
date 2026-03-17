package algoliautil

import (
	"errors"
	"net/http"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/errs"
	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

func IsRetryableError(err error) bool {
	if IsNotFoundError(err) {
		return true
	}
	if errors.Is(err, errs.ErrNoMoreHostToTry) {
		return true
	}
	return false
}

func IsNotFoundError(err error) bool {
	var apiErr *search.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == http.StatusNotFound
	}
	return false
}
