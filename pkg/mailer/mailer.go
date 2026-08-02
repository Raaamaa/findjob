package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
)

type Mailer interface {
	SendEmail(to string, subject string, body string, attachmentPath string) error
}

type SMTPMailer struct {
	Host       string
	Port       int
	Username   string
	Password   string
	SenderName string
}

// NewSMTPMailer creates a new SMTP-based Mailer instance.
func NewSMTPMailer(host string, port int, username string, password string, senderName string) *SMTPMailer {
	return &SMTPMailer{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		SenderName: senderName,
	}
}

// SendEmail sends a raw SMTP email with optional PDF attachment.
func (m *SMTPMailer) SendEmail(to string, subject string, body string, attachmentPath string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	buf := bytes.NewBuffer(nil)

	fromHeader := fmt.Sprintf("%s <%s>", m.SenderName, m.Username)
	if m.SenderName == "" {
		fromHeader = m.Username
	}
	buf.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	boundary := "jobber_boundary_987654321"
	dDash := "-" + "-"

	if attachmentPath != "" {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("%s%s\r\n", dDash, boundary))
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(body)
		buf.WriteString("\r\n")

		fileBytes, err := os.ReadFile(attachmentPath)
		if err != nil {
			return fmt.Errorf("failed to read attachment file: %w", err)
		}
		fileName := filepath.Base(attachmentPath)
		mimeType := mime.TypeByExtension(filepath.Ext(fileName))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		buf.WriteString(fmt.Sprintf("%s%s\r\n", dDash, boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", mimeType, fileName))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", fileName))
		buf.WriteString("\r\n")

		b64Content := base64.StdEncoding.EncodeToString(fileBytes)
		for i := 0; i < len(b64Content); i += 76 {
			end := i + 76
			if end > len(b64Content) {
				end = len(b64Content)
			}
			buf.WriteString(b64Content[i:end])
			buf.WriteString("\r\n")
		}

		buf.WriteString(fmt.Sprintf("%s%s%s\r\n", dDash, boundary, dDash))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(body)
	}

	err := smtp.SendMail(addr, auth, m.Username, []string{to}, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send smtp mail: %w", err)
	}

	return nil
}
