package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryDuplicateCheck(t *testing.T) {
	entries := []Entry{
		{
			ImageHash: "hash123",
			Company:   "Google",
			Position:  "Software Engineer",
			Email:     "jobs@google.com",
			AppliedAt: time.Now(),
		},
	}

	// 1. Check exact match of image hash
	hasImage, hasContact := CheckDuplicate(entries, "hash123", "other@google.com", "Designer")
	if !hasImage {
		t.Error("expected hasImage to be true")
	}
	if hasContact {
		t.Error("expected hasContact to be false")
	}

	// 2. Check exact match of contact email and position (case-insensitive)
	hasImage, hasContact = CheckDuplicate(entries, "hash456", "JOBS@google.com", "software engineer")
	if hasImage {
		t.Error("expected hasImage to be false")
	}
	if !hasContact {
		t.Error("expected hasContact to be true")
	}

	// 3. No match
	hasImage, hasContact = CheckDuplicate(entries, "hash456", "other@google.com", "Product Manager")
	if hasImage || hasContact {
		t.Error("expected hasImage and hasContact to be false")
	}
}

func TestLoadAndSaveHistory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jobber_history_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	historyFile := filepath.Join(tempDir, "history.json")

	// Load non-existent file (should return empty list, no error)
	entries, err := LoadHistory(historyFile)
	if err != nil {
		t.Fatalf("expected no error loading non-existent file, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(entries))
	}

	// Add entry
	err = AddEntry(historyFile, "Test Company", "Developer", "hr@test.com", "")
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	entries, err = LoadHistory(historyFile)
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Company != "Test Company" || entries[0].Position != "Developer" || entries[0].Email != "hr@test.com" {
		t.Errorf("incorrect entry loaded: %+v", entries[0])
	}
}
