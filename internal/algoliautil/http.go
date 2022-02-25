package algoliautil

import (
	"net/http"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/transport"
)

type DebugRequester struct {
	Client *http.Client
}

func NewDebugRequester() *DebugRequester {
	return &DebugRequester{
		Client: transport.DefaultHTTPClient(),
	}
}

func (d *DebugRequester) Request(req *http.Request) (*http.Response, error) {
	return d.Client.Do(req)
}
