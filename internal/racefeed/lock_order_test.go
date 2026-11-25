package racefeed

import (
	"sync"
	"testing"
	"time"
)

// Concurrent UpdateInfo + Snapshot used to AB-BA deadlock (detector then e.mu vs e.mu then detector).
func TestUpdateInfoSnapshotLockOrder(t *testing.T) {
	e := New()
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			e.UpdateInfo(i%13, "Driver", "car01", "Car", false)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			_ = e.Snapshot()
		}
	}()
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("UpdateInfo + Snapshot deadlocked (lock order)")
	}
}
