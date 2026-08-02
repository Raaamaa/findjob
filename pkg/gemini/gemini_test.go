package gemini

import (
	"encoding/json"
	"testing"
)

func TestParseResponseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid json",
			input: `{
				"company": "Example Co",
				"position": "Go Developer",
				"email": "hr@example.com",
				"subject": "Application for Go Developer",
				"email_draft": "Hello, this is my draft email."
			}`,
			wantErr: false,
		},
		{
			name: "invalid json syntax",
			input: `{
				"company": "Example Co",
				"position": "Go Developer",
				"email": "hr@example.com",
				"subject": "Application for Go Developer"
				"email_draft": "Hello, this is my draft email."
			}`, // missing comma
			wantErr: true,
		},
		{
			name: "extra fields",
			input: `{
				"company": "Example Co",
				"position": "Go Developer",
				"email": "hr@example.com",
				"subject": "Application for Go Developer",
				"email_draft": "Hello.",
				"random_field": "extra value"
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var details JobDetails
			err := json.Unmarshal([]byte(tt.input), &details)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if details.Company != "Example Co" {
					t.Errorf("expected company Example Co, got %s", details.Company)
				}
				if details.Position != "Go Developer" {
					t.Errorf("expected position Go Developer, got %s", details.Position)
				}
			}
		})
	}
}

func TestDetectMIMEType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"IMAGE.PNG", "image/png"},
		{"doc.webp", "image/webp"},
		{"doc.pdf", ""},
	}

	for _, tt := range tests {
		res := detectMIMEType(tt.path)
		if res != tt.expected {
			t.Errorf("detectMIMEType(%s) = %s, expected %s", tt.path, res, tt.expected)
		}
	}
}
