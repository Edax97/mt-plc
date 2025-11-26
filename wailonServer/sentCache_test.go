package wailonServer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to temp file path without creating it
func tempFilePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "sentcache.gob")
}

func TestNewSentCache_EmptyWhenMissing(t *testing.T) {
	path := tempFilePath(t)
	cache := NewSentCache(path)
	if cache == nil {
		t.Fatalf("NewSentCache returned nil")
	}
	if cache.diskPath != path {
		t.Fatalf("diskPath mismatch: got %s want %s", cache.diskPath, path)
	}
	if cache.sentMap == nil {
		t.Fatalf("sentMap is nil")
	}
	if len(cache.sentMap) != 0 {
		t.Fatalf("expected empty sentMap for missing file, got %d entries", len(cache.sentMap))
	}
}

func TestUpdateSentAndPersistence(t *testing.T) {
	path := tempFilePath(t)

	c1 := NewSentCache(path)
	imei := "12345"
	t1 := time.Now().UTC().Truncate(time.Second) // truncate to stable seconds

	if ok := c1.UpdateSent(imei, t1); !ok {
		t.Fatalf("updateSent returned false")
	}

	// Load again and verify the time is persisted
	c2 := NewSentCache(path)

	// Direct map check
	got, ok := c2.sentMap[imei]
	if !ok {
		t.Fatalf("persisted imei not found in loaded cache")
	}
	if !got.Equal(t1) {
		t.Fatalf("loaded time mismatch: got %v want %v", got, t1)
	}

	// HasSent semantics: true only when query time is strictly before stored time
	if !c2.HasSent(imei, t1.Add(-1*time.Second)) {
		t.Fatalf("HasSent should be true for time before stored time")
	}
	if c2.HasSent(imei, t1) {
		t.Fatalf("HasSent should be false for time equal to stored time")
	}
	if c2.HasSent(imei, t1.Add(1*time.Second)) {
		t.Fatalf("HasSent should be false for time after stored time")
	}
}

func TestCorruptedFileHandledGracefully(t *testing.T) {
	path := tempFilePath(t)

	// Create file with garbage
	if err := os.WriteFile(path, []byte("not-a-gob"), 0o644); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	// Should not panic and should initialize empty map
	c := NewSentCache(path)
	if c.sentMap == nil {
		t.Fatalf("sentMap is nil")
	}
	if len(c.sentMap) != 0 {
		t.Fatalf("expected empty map after failed load, got %d entries", len(c.sentMap))
	}
}

func TestSaveCreatesNonEmptyFile(t *testing.T) {
	path := tempFilePath(t)
	c := NewSentCache(path)

	c.UpdateSent("A", time.Now())

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}
	if fi.Size() <= 0 {
		t.Fatalf("expected cache file to be non-empty, size=%d", fi.Size())
	}
}
