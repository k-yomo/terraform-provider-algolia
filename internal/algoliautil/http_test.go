package algoliautil

import (
	"reflect"
	"testing"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/transport"
)

func TestNewDebugRequester(t *testing.T) {
	t.Parallel()

	got := NewDebugRequester()
	want := &DebugRequester{
		Client: transport.DefaultHTTPClient(),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewDebugRequester() = %v, want %v", got, want)
	}
}
