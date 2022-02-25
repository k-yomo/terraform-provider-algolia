package algoliautil

import (
	"reflect"
	"testing"
)

func TestNewDebugRequester(t *testing.T) {
	t.Parallel()

	got := NewDebugRequester()
	// assert not nil
	if reflect.DeepEqual(got, nil) {
		t.Errorf("NewDebugRequester() = %v, want %v", got, nil)
	}
}
