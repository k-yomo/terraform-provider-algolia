package algoliautil

import (
	"github.com/algolia/algoliasearch-client-go/v3/algolia/errs"
	"net/http"
)

func IsAlgoliaNotFoundError(err error) bool {
	if algoliaErr, ok := errs.IsAlgoliaErr(err); ok && algoliaErr.Status == http.StatusNotFound {
		return true
	}
	return false
}
