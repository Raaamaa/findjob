package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jobber_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	envFile := filepath.Join(tempDir, ".env")
	content := `
# Comment line
TELEGRAM_BOT_TOKEN=test_bot_token
GEMINI_API_KEY=test_api_key
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user@example.com
SMTP_PASS=password123
DEFAULT_SENDER_NAME=Test Name
DEV_CV_PATH=cv-dev.pdf
DEV_CV_SUMMARY_PATH=cv-dev.md
FNB_CV_PATH=cv-fnb.pdf
FNB_CV_SUMMARY_PATH=cv-fnb.md
ALLOWED_TELEGRAM_USERNAME=test_user
`
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test env: %v", err)
	}

	cfg, err := LoadConfig(envFile)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.TelegramBotToken != "test_bot_token" {
		t.Errorf("expected TelegramBotToken test_bot_token, got %s", cfg.TelegramBotToken)
	}
	if cfg.GeminiAPIKey != "test_api_key" {
		t.Errorf("expected GeminiAPIKey test_api_key, got %s", cfg.GeminiAPIKey)
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("expected SMTPHost smtp.example.com, got %s", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 587 {
		t.Errorf("expected SMTPPort 587, got %d", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "user@example.com" {
		t.Errorf("expected SMTPUser user@example.com, got %s", cfg.SMTPUser)
	}
	if cfg.SMTPPass != "password123" {
		t.Errorf("expected SMTPPass password123, got %s", cfg.SMTPPass)
	}
	if cfg.DefaultSenderName != "Test Name" {
		t.Errorf("expected DefaultSenderName Test Name, got %s", cfg.DefaultSenderName)
	}
	if cfg.DevCVPath != "cv-dev.pdf" {
		t.Errorf("expected DevCVPath cv-dev.pdf, got %s", cfg.DevCVPath)
	}
	if cfg.DevCVSummaryPath != "cv-dev.md" {
		t.Errorf("expected DevCVSummaryPath cv-dev.md, got %s", cfg.DevCVSummaryPath)
	}
	if cfg.FnBCVPath != "cv-fnb.pdf" {
		t.Errorf("expected FnBCVPath cv-fnb.pdf, got %s", cfg.FnBCVPath)
	}
	if cfg.FnBCVSummaryPath != "cv-fnb.md" {
		t.Errorf("expected FnBCVSummaryPath cv-fnb.md, got %s", cfg.FnBCVSummaryPath)
	}
	if cfg.AllowedTelegramUsername != "test_user" {
		t.Errorf("expected AllowedTelegramUsername test_user, got %s", cfg.AllowedTelegramUsername)
	}
}

func TestLoadConfig_MissingKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jobber_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	envFile := filepath.Join(tempDir, ".env")
	content := `
GEMINI_API_KEY=test_api_key
`
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test env: %v", err)
	}

	_, err = LoadConfig(envFile)
	if err == nil {
		t.Error("expected error due to missing keys, got nil")
	}
}
