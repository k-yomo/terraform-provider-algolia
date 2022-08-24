package mutex

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// KV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
type KV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// NewKV returns a properly initialized KV
func NewKV() *KV {
	return &KV{
		store: make(map[string]*sync.Mutex),
	}
}

// Lock the mutex for the given key. Caller is responsible for calling Unlock
// for the same key
func (m *KV) Lock(ctx context.Context, key string) {
	tflog.Trace(ctx, "Locking", map[string]interface{}{"key": key})
	m.get(key).Lock()
	tflog.Trace(ctx, "Locked", map[string]interface{}{"key": key})
}

// Unlock the mutex for the given key. Caller must have called Lock for the same key first
func (m *KV) Unlock(ctx context.Context, key string) {
	tflog.Trace(ctx, "Unlocking", map[string]interface{}{"key": key})
	m.get(key).Unlock()
	tflog.Trace(ctx, "Unlocked", map[string]interface{}{"key": key})
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *KV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}
