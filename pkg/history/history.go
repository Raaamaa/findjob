package history

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Entry struct {
	ImageHash string    `json:"image_hash"`
	Company   string    `json:"company"`
	Position  string    `json:"position"`
	Email     string    `json:"email"`
	AppliedAt time.Time `json:"applied_at"`
}

// ComputeSHA256 calculates the SHA256 checksum of a file.
func ComputeSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ComputeSHA256Bytes calculates the SHA256 checksum of a byte slice.
func ComputeSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

// LoadHistory reads history entries from the json file.
func LoadHistory(filePath string) ([]Entry, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []Entry{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	if len(data) == 0 {
		return []Entry{}, nil
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse history json: %w", err)
	}

	return entries, nil
}

// SaveHistory writes the history entries to the json file.
func SaveHistory(filePath string, entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// CheckDuplicate checks if the image hash or email/position combination already exists.
func CheckDuplicate(entries []Entry, imageHash string, email string, position string) (hasImage bool, hasContact bool) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanPosition := strings.ToLower(strings.TrimSpace(position))

	for _, entry := range entries {
		if imageHash != "" && entry.ImageHash == imageHash {
			hasImage = true
		}
		if cleanEmail != "" && cleanPosition != "" &&
			strings.ToLower(entry.Email) == cleanEmail &&
			strings.ToLower(entry.Position) == cleanPosition {
			hasContact = true
		}
	}
	return hasImage, hasContact
}

// AddEntry records a new transaction in the history database file.
func AddEntry(filePath string, company string, position string, email string, imageHash string) error {
	entries, err := LoadHistory(filePath)
	if err != nil {
		return err
	}

	newEntry := Entry{
		ImageHash: imageHash,
		Company:   company,
		Position:  position,
		Email:     email,
		AppliedAt: time.Now(),
	}

	entries = append(entries, newEntry)
	return SaveHistory(filePath, entries)
}
