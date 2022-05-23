package mutex

import (
	"sync"
	"testing"
	"time"
)

func TestKV(t *testing.T) {
	t.Parallel()

	mutexKV := NewKV()
	start := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		mutexKV.Lock("test")
		time.Sleep(100 * time.Millisecond)
		mutexKV.Unlock("test")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		mutexKV.Lock("test")
		time.Sleep(100 * time.Millisecond)
		mutexKV.Unlock("test")
	}()
	wg.Wait()

	if elapsed := time.Since(start); elapsed < 200*time.Millisecond {
		t.Errorf("TestKV() elapsed time = %v, want %v", elapsed, 200*time.Millisecond)
	}
}
