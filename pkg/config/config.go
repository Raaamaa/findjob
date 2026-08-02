package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramBotToken  string
	GeminiAPIKey      string
	GeminiModel       string
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	DefaultSenderName string
	DevCVPath         string
	DevCVSummaryPath  string
	FnBCVPath         string
	FnBCVSummaryPath  string
}

// LoadConfig parses the configuration key-value pairs from the environment file.
func LoadConfig(envPath string) (*Config, error) {
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", envPath)
	}

	file, err := os.Open(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}
		envVars[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	model := envVars["GEMINI_MODEL"]
	if model == "" {
		model = "gemini-3.5-flash-lite"
	}

	cfg := &Config{
		TelegramBotToken:  envVars["TELEGRAM_BOT_TOKEN"],
		GeminiAPIKey:      envVars["GEMINI_API_KEY"],
		GeminiModel:       model,
		SMTPHost:          envVars["SMTP_HOST"],
		SMTPUser:          envVars["SMTP_USER"],
		SMTPPass:          envVars["SMTP_PASS"],
		DefaultSenderName: envVars["DEFAULT_SENDER_NAME"],
		DevCVPath:         envVars["DEV_CV_PATH"],
		DevCVSummaryPath:  envVars["DEV_CV_SUMMARY_PATH"],
		FnBCVPath:         envVars["FNB_CV_PATH"],
		FnBCVSummaryPath:  envVars["FNB_CV_SUMMARY_PATH"],
	}

	portStr := envVars["SMTP_PORT"]
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %s", portStr)
		}
		cfg.SMTPPort = port
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required in %s", envPath)
	}
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required in %s", envPath)
	}
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is required in %s", envPath)
	}
	if cfg.SMTPPort == 0 {
		return nil, fmt.Errorf("SMTP_PORT is required in %s", envPath)
	}
	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP_USER is required in %s", envPath)
	}
	if cfg.SMTPPass == "" {
		return nil, fmt.Errorf("SMTP_PASS is required in %s", envPath)
	}
	if cfg.DevCVPath == "" {
		return nil, fmt.Errorf("DEV_CV_PATH is required in %s", envPath)
	}
	if cfg.DevCVSummaryPath == "" {
		return nil, fmt.Errorf("DEV_CV_SUMMARY_PATH is required in %s", envPath)
	}
	if cfg.FnBCVPath == "" {
		return nil, fmt.Errorf("FNB_CV_PATH is required in %s", envPath)
	}
	if cfg.FnBCVSummaryPath == "" {
		return nil, fmt.Errorf("FNB_CV_SUMMARY_PATH is required in %s", envPath)
	}

	return cfg, nil
}

// CheckPermissions alerts the user if the config file has insecure permission settings.
func CheckPermissions(envPath string) {
	info, err := os.Stat(envPath)
	if err != nil {
		return
	}

	mode := info.Mode().Perm()
	if mode&0o077 != 0 {
		fmt.Printf("WARNING: file %s has insecure permissions: %04o. Recommended permissions: 0600 (chmod 600 %s)\n", envPath, mode, envPath)
	}
}
