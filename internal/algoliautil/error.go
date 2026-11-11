package algoliautil

import (
	"errors"
	"net/http"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/errs"
	suggestions "github.com/algolia/algoliasearch-client-go/v4/algolia/query-suggestions"
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

func IsForbiddenError(err error) bool {
	return apiErrorStatus(err) == http.StatusForbidden
}

func IsNotFoundError(err error) bool {
	return apiErrorStatus(err) == http.StatusNotFound
}

// apiErrorStatus returns the HTTP status of an Algolia API error, or 0 if err is not one.
// Each v4 API client package defines its own APIError type, so all used clients must be checked.
func apiErrorStatus(err error) int {
	var searchErr *search.APIError
	if errors.As(err, &searchErr) {
		return searchErr.Status
	}
	var suggestionsErr *suggestions.APIError
	if errors.As(err, &suggestionsErr) {
		return suggestionsErr.Status
	}
	return 0
}
