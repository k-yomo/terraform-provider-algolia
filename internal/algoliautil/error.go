package algoliautil

import (
	"net/http"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
)

func IsRetryableError(err error) bool {
	if IsNotFoundError(err) {
		return true
	}
	if _, ok := err.(*errs.NoMoreHostToTryErr); ok {
		return true
	}
	return false
}

func IsNotFoundError(err error) bool {
	_, ok := errs.IsAlgoliaErrWithCode(err, http.StatusNotFound)
	return ok
}
