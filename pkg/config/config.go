package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramBotToken        string
	GeminiAPIKey            string
	GeminiModel             string
	SMTPHost                string
	SMTPPort                int
	SMTPUser                string
	SMTPPass                string
	DefaultSenderName       string
	CVPath                  string
	CVSummaryPath           string
	AllowedTelegramUsername string
}

// LoadConfig parses the configuration key-value pairs from the environment file or system environment variables.
func LoadConfig(envPath string) (*Config, error) {
	envVars := make(map[string]string)

	if _, err := os.Stat(envPath); err == nil {
		file, err := os.Open(envPath)
		if err == nil {
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
			file.Close()
		}
	}

	getEnv := func(key string) string {
		if val, ok := envVars[key]; ok && val != "" {
			return val
		}
		return os.Getenv(key)
	}

	model := getEnv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-flash-latest"
	}

	cfg := &Config{
		TelegramBotToken:        getEnv("TELEGRAM_BOT_TOKEN"),
		GeminiAPIKey:            getEnv("GEMINI_API_KEY"),
		GeminiModel:             model,
		SMTPHost:                getEnv("SMTP_HOST"),
		SMTPUser:                getEnv("SMTP_USER"),
		SMTPPass:                getEnv("SMTP_PASS"),
		DefaultSenderName:       getEnv("DEFAULT_SENDER_NAME"),
		CVPath:                  getEnv("CV_PATH"),
		CVSummaryPath:           getEnv("CV_SUMMARY_PATH"),
		AllowedTelegramUsername: strings.TrimPrefix(getEnv("ALLOWED_TELEGRAM_USERNAME"), "@"),
	}

	portStr := getEnv("SMTP_PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP_PORT: %s", portStr)
		}
		cfg.SMTPPort = port
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}
	if cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is required")
	}
	if cfg.SMTPPort == 0 {
		return nil, fmt.Errorf("SMTP_PORT is required")
	}
	if cfg.SMTPUser == "" {
		return nil, fmt.Errorf("SMTP_USER is required")
	}
	if cfg.SMTPPass == "" {
		return nil, fmt.Errorf("SMTP_PASS is required")
	}
	if cfg.CVPath == "" {
		return nil, fmt.Errorf("CV_PATH is required")
	}
	if cfg.CVSummaryPath == "" {
		return nil, fmt.Errorf("CV_SUMMARY_PATH is required")
	}
	if cfg.AllowedTelegramUsername == "" {
		return nil, fmt.Errorf("ALLOWED_TELEGRAM_USERNAME is required")
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
