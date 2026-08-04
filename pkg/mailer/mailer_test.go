package mailer

import "testing"

type MockMailer struct {
	To             string
	Subject        string
	Body           string
	AttachmentPath string
	SendErr        error
	Called         bool
}

func (m *MockMailer) SendEmail(to string, subject string, body string, attachmentPath string) error {
	m.To = to
	m.Subject = subject
	m.Body = body
	m.AttachmentPath = attachmentPath
	m.Called = true
	return m.SendErr
}

func TestMockMailer(t *testing.T) {
	mock := &MockMailer{}
	err := mock.SendEmail("test@example.com", "Test Subject", "Test Body", "cv.pdf")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !mock.Called {
		t.Error("expected MockMailer to be marked called")
	}
	if mock.To != "test@example.com" {
		t.Errorf("expected To test@example.com, got %s", mock.To)
	}
	if mock.Subject != "Test Subject" {
		t.Errorf("expected Subject 'Test Subject', got '%s'", mock.Subject)
	}
	if mock.Body != "Test Body" {
		t.Errorf("expected Body 'Test Body', got '%s'", mock.Body)
	}
	if mock.AttachmentPath != "cv.pdf" {
		t.Errorf("expected AttachmentPath cv.pdf, got %s", mock.AttachmentPath)
	}
}

func TestValidateEmail(t *testing.T) {
	valid := []string{"test@example.com", "user.name+tag@domain.co.id"}
	invalid := []string{"", "invalid-email", "user@.com", "@domain.com"}

	for _, e := range valid {
		if err := ValidateEmail(e); err != nil {
			t.Errorf("expected valid email for '%s', got error: %v", e, err)
		}
	}

	for _, e := range invalid {
		if err := ValidateEmail(e); err == nil {
			t.Errorf("expected error for invalid email '%s', got nil", e)
		}
	}
}
