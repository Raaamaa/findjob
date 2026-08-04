package mailer

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
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

// ValidateEmail checks if the given email address has a valid syntax.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email address is empty")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address '%s': %w", email, err)
	}
	return nil
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
	if err := ValidateEmail(to); err != nil {
		return err
	}

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
		buf.WriteString(toCRLF(body))
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
		buf.WriteString(toCRLF(body))
	}

	if m.Port == 465 {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         m.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return fmt.Errorf("failed to connect via TLS to %s: %w", addr, err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, m.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err = client.Mail(m.Username); err != nil {
			return fmt.Errorf("SMTP MAIL command failed: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP RCPT command failed: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("SMTP DATA command failed: %w", err)
		}

		_, err = w.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write SMTP payload: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close SMTP writer: %w", err)
		}

		return nil
	}

	err := smtp.SendMail(addr, auth, m.Username, []string{to}, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send smtp mail: %w", err)
	}

	return nil
}

// toCRLF normalizes line endings in a string to CRLF (\r\n) as required by SMTP specs.
func toCRLF(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	return strings.ReplaceAll(input, "\n", "\r\n")
}
