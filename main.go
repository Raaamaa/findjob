package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"jobber/pkg/config"
	"jobber/pkg/gemini"
	"jobber/pkg/history"
	"jobber/pkg/mailer"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type PendingJob struct {
	Company    string
	Position   string
	Email      string
	Subject    string
	EmailDraft string
	CVPath     string
	ImageHash  string
}

var (
	pendingJobs = make(map[int64]*PendingJob)
	pendingMu   sync.Mutex
)

func main() {
	envPath := ".env"

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		log.Fatalf("Error: Berkas konfigurasi %s tidak ditemukan.", envPath)
	}

	config.CheckPermissions(envPath)

	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		log.Fatalf("Error: Gagal memuat konfigurasi: %v", err)
	}

	if _, err := os.Stat(cfg.CVPath); os.IsNotExist(err) {
		log.Fatalf("Error: Berkas CV PDF '%s' tidak ditemukan.", cfg.CVPath)
	}
	if _, err := os.Stat(cfg.CVSummaryPath); os.IsNotExist(err) {
		log.Fatalf("Error: Berkas Ringkasan CV '%s' tidak ditemukan.", cfg.CVSummaryPath)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram Bot: %v", err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	ctx := context.Background()
	geminiSvc, err := gemini.NewService(ctx, cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini Service: %v", err)
	}

	for update := range updates {
		if update.Message != nil {
			if update.Message.From == nil || strings.ToLower(update.Message.From.UserName) != strings.ToLower(cfg.AllowedTelegramUsername) {
				continue
			}
			handleMessage(ctx, bot, update.Message, geminiSvc, cfg)
		} else if update.CallbackQuery != nil {
			if update.CallbackQuery.From == nil || strings.ToLower(update.CallbackQuery.From.UserName) != strings.ToLower(cfg.AllowedTelegramUsername) {
				continue
			}
			handleCallbackQuery(bot, update.CallbackQuery, cfg)
		}
	}
}

func handleMessage(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message, geminiSvc *gemini.Service, cfg *config.Config) {
	chatID := msg.Chat.ID

	// Handle clearing session
	if msg.Text == ".clear" || msg.Text == "/clear" {
		pendingMu.Lock()
		_, exists := pendingJobs[chatID]
		if exists {
			delete(pendingJobs, chatID)
		}
		pendingMu.Unlock()

		bot.Send(tgbotapi.NewMessage(chatID, "Sesi aktif berhasil dibersihkan."))
		return
	}

	// Handle editing session details
	if strings.HasPrefix(msg.Text, "/edit ") || strings.HasPrefix(msg.Text, ".edit ") {
		parts := strings.SplitN(msg.Text, " ", 3)
		if len(parts) < 3 {
			bot.Send(tgbotapi.NewMessage(chatID, "Format salah. Gunakan: /edit <company/position/email/subject/body> <nilai_baru>"))
			return
		}
		field := strings.ToLower(parts[1])
		value := parts[2]

		pendingMu.Lock()
		pending, exists := pendingJobs[chatID]
		if !exists {
			pendingMu.Unlock()
			bot.Send(tgbotapi.NewMessage(chatID, "Tidak ada transaksi aktif untuk disunting. Silakan kirim gambar lowongan terlebih dahulu."))
			return
		}

		switch field {
		case "company":
			pending.Company = value
		case "position":
			pending.Position = value
		case "email":
			pending.Email = value
		case "subject":
			pending.Subject = value
		case "body":
			pending.EmailDraft = value
		default:
			pendingMu.Unlock()
			bot.Send(tgbotapi.NewMessage(chatID, "Bidang tidak dikenal. Pilih salah satu: company, position, email, subject, body"))
			return
		}

		// Copy fields to local variables to safely construct the preview outside the lock
		company := pending.Company
		position := pending.Position
		email := pending.Email
		subject := pending.Subject
		draft := pending.EmailDraft
		pendingMu.Unlock()

		var previewBuilder strings.Builder
		if err := mailer.ValidateEmail(email); err != nil {
			previewBuilder.WriteString("⚠️ PERINGATAN: Format email tujuan belum valid. Gunakan '/edit email <alamat_email>' untuk mengoreksi.\n\n")
		}
		previewBuilder.WriteString(fmt.Sprintf("=== DETAIL LOWONGAN (UPDATED) ===\nPerusahaan: %s\nPosisi: %s\nEmail Tujuan: %s\nSubjek: %s\n=======================\n\n*** DRAFT EMAIL LAMARAN ***\n%s\n***************************",
			company, position, email, subject, draft))

		replyMsg := tgbotapi.NewMessage(chatID, previewBuilder.String())
		replyMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Accept (Kirim)", "accept"),
				tgbotapi.NewInlineKeyboardButtonData("Deny (Batal)", "deny"),
			),
		)
		bot.Send(replyMsg)
		return
	}

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			reply := "Halo! Selamat datang di Jobber Bot.\n\nKirimkan gambar/screenshot lowongan kerja Anda ke chat ini, dan saya akan mengekstrak detail lowongan, merumuskan draf email lamaran secara otomatis sesuai CV Anda, dan mengirimkannya setelah Anda melakukan konfirmasi.\n\nAnda dapat mengoreksi hasil ekstraksi sebelum mengirim menggunakan perintah:\n- /edit company <nama_perusahaan>\n- /edit position <nama_posisi>\n- /edit email <email_baru>\n- /edit subject <subjek_baru>\n- /edit body <draf_baru>\n\nGunakan perintah /clear atau .clear untuk menghapus sesi aktif Anda."
			bot.Send(tgbotapi.NewMessage(chatID, reply))
		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Perintah tidak dikenal. Kirimkan gambar lowongan kerja untuk memulai."))
		}
		return
	}

	if len(msg.Photo) > 0 {
		statusMsg, _ := bot.Send(tgbotapi.NewMessage(chatID, "Menerima gambar lowongan kerja. Sedang memproses analisis..."))

		photo := msg.Photo[len(msg.Photo)-1]
		fileURL, err := bot.GetFileDirectURL(photo.FileID)
		if err != nil {
			updateStatus(bot, chatID, statusMsg.MessageID, fmt.Sprintf("Error: Gagal mendapatkan URL file dari Telegram: %v", err))
			return
		}

		updateStatus(bot, chatID, statusMsg.MessageID, "Mengunduh gambar ke memori...")
		imageBytes, mimeType, err := downloadPhoto(fileURL)
		if err != nil {
			updateStatus(bot, chatID, statusMsg.MessageID, fmt.Sprintf("Error: Gagal mengunduh gambar: %v", err))
			return
		}

		imgHash := history.ComputeSHA256Bytes(imageBytes)

		historyPath := "history.json"
		entries, err := history.LoadHistory(historyPath)
		var warnText []string
		if err == nil {
			hasImage, _ := history.CheckDuplicate(entries, imgHash, "", "")
			if hasImage {
				warnText = append(warnText, "WARNING: Gambar lowongan ini terdeteksi sudah pernah Anda proses sebelumnya.")
			}
		}

		updateStatus(bot, chatID, statusMsg.MessageID, "Membaca data CV dan menganalisis lowongan kerja dengan AI...")

		cvSummaryBytes, err := os.ReadFile(cfg.CVSummaryPath)
		if err != nil {
			updateStatus(bot, chatID, statusMsg.MessageID, fmt.Sprintf("Error: Gagal membaca file ringkasan CV: %v", err))
			return
		}

		details, err := geminiSvc.ExtractJobDetails(ctx, imageBytes, mimeType, string(cvSummaryBytes))
		if err != nil {
			updateStatus(bot, chatID, statusMsg.MessageID, fmt.Sprintf("Error: Gagal menganalisis lowongan via Gemini: %v", err))
			return
		}

		if err := mailer.ValidateEmail(details.Email); err != nil {
			warnText = append(warnText, "⚠️ PERINGATAN: Email penerima tidak terdeteksi atau tidak valid pada gambar lowongan. Gunakan perintah '/edit email <alamat_email>' untuk melengkapi.")
		}

		_, hasContact := history.CheckDuplicate(entries, "", details.Email, details.Position)
		if hasContact {
			warnText = append(warnText, fmt.Sprintf("WARNING: Anda terdeteksi sudah pernah melamar posisi '%s' ke email '%s' sebelumnya.", details.Position, details.Email))
		}

		pendingMu.Lock()
		pendingJobs[chatID] = &PendingJob{
			Company:    details.Company,
			Position:   details.Position,
			Email:      details.Email,
			Subject:    details.Subject,
			EmailDraft: details.EmailDraft,
			CVPath:     cfg.CVPath,
			ImageHash:  imgHash,
		}
		pendingMu.Unlock()

		bot.Send(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID))

		var previewBuilder strings.Builder
		if len(warnText) > 0 {
			previewBuilder.WriteString(strings.Join(warnText, "\n"))
			previewBuilder.WriteString("\n\n")
		}
		previewBuilder.WriteString(fmt.Sprintf("=== DETAIL LOWONGAN ===\nPerusahaan: %s\nPosisi: %s\nEmail Tujuan: %s\nSubjek: %s\n=======================\n\n*** DRAFT EMAIL LAMARAN ***\n%s\n***************************",
			details.Company, details.Position, details.Email, details.Subject, details.EmailDraft))

		replyMsg := tgbotapi.NewMessage(chatID, previewBuilder.String())
		replyMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Accept (Kirim)", "accept"),
				tgbotapi.NewInlineKeyboardButtonData("Deny (Batal)", "deny"),
			),
		)
		bot.Send(replyMsg)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "Kirimkan gambar/screenshot lowongan kerja untuk memulai pemrosesan."))
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, cfg *config.Config) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	callback := tgbotapi.NewCallback(cb.ID, "")
	bot.Request(callback)

	pendingMu.Lock()
	pendingPtr, exists := pendingJobs[chatID]
	var pending PendingJob
	if exists && pendingPtr != nil {
		pending = *pendingPtr
	}
	pendingMu.Unlock()

	if !exists {
		bot.Send(tgbotapi.NewMessage(chatID, "Tidak ada transaksi aktif yang sedang menunggu konfirmasi Anda."))
		return
	}

	switch cb.Data {
	case "accept":
		statusMsg, _ := bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Mengirim email lamaran ke %s...", pending.Email)))

		mailSvc := mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.DefaultSenderName)
		err := mailSvc.SendEmail(pending.Email, pending.Subject, pending.EmailDraft, pending.CVPath)
		if err != nil {
			bot.Send(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID))
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Gagal mengirim email: %v", err)))
			return
		}

		bot.Send(tgbotapi.NewDeleteMessage(chatID, statusMsg.MessageID))

		historyPath := "history.json"
		err = history.AddEntry(historyPath, pending.Company, pending.Position, pending.Email, pending.ImageHash)
		if err != nil {
			log.Printf("Warning: Failed to log history: %v", err)
		}

		edit := tgbotapi.NewEditMessageText(chatID, messageID, fmt.Sprintf("=== STATUS: EMAIL TERKIRIM ===\n\nLamaran untuk posisi '%s' di '%s' telah berhasil dikirimkan ke %s.", pending.Position, pending.Company, pending.Email))
		bot.Send(edit)

		pendingMu.Lock()
		delete(pendingJobs, chatID)
		pendingMu.Unlock()

	case "deny":
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "=== STATUS: DITOLAK / DIBATALKAN ===")
		bot.Send(edit)

		pendingMu.Lock()
		delete(pendingJobs, chatID)
		pendingMu.Unlock()
	}
}

func updateStatus(bot *tgbotapi.BotAPI, chatID int64, messageID int, newText string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, newText)
	bot.Send(edit)
}

func downloadPhoto(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		detected := http.DetectContentType(data)
		if detected != "application/octet-stream" {
			mimeType = detected
		} else {
			if strings.HasSuffix(strings.ToLower(url), ".png") {
				mimeType = "image/png"
			} else if strings.HasSuffix(strings.ToLower(url), ".webp") {
				mimeType = "image/webp"
			} else {
				mimeType = "image/jpeg"
			}
		}
	}

	return data, mimeType, nil
}
