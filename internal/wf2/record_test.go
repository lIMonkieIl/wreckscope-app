package wf2

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCaptureReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte{0, 1, 2, 3, 4, 255}
	if err := WriteCapture(f, payload, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	count := 0
	if err := Replay(path, 1, func(v []byte) {
		count++
		if !bytes.Equal(payload, v) {
			t.Errorf("payload mismatch %v", v)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d packets", count)
	}
}
