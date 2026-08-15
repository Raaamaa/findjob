package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

type JobDetails struct {
	Company    string `json:"company"`
	Position   string `json:"position"`
	Email      string `json:"email"`
	Subject    string `json:"subject"`
	EmailDraft string `json:"email_draft"`
}

type Service struct {
	client *genai.Client
	model  string
}

// NewService creates a new Gemini service wrapper client.
func NewService(ctx context.Context, apiKey string, model string) (*Service, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &Service{
		client: client,
		model:  model,
	}, nil
}

// ExtractJobDetails extracts relevant info from the job ad image and generates an email draft using a single CV summary.
func (s *Service) ExtractJobDetails(ctx context.Context, imageBytes []byte, mimeType string, cvSummary string) (*JobDetails, error) {
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"company":     {Type: genai.TypeString, Description: "Nama perusahaan yang membuka lowongan"},
			"position":    {Type: genai.TypeString, Description: "Posisi pekerjaan yang dilamar"},
			"email":       {Type: genai.TypeString, Description: "Alamat email penerima lamaran. Jika tidak ada di gambar, isi string kosong."},
			"subject":     {Type: genai.TypeString, Description: "Subjek email lamaran yang profesional"},
			"email_draft": {Type: genai.TypeString, Description: "Draft email lamaran / cover letter dalam bahasa Indonesia yang disesuaikan dengan deskripsi pekerjaan pada gambar lowongan dan menggunakan informasi dari data CV pelamar. Tulis secara sopan dan profesional. Jangan ada placeholder yang belum terisi."},
		},
		Required: []string{"company", "position", "email", "subject", "email_draft"},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	prompt := fmt.Sprintf(`Analisis gambar lowongan kerja ini dan ekstrak informasi yang diperlukan untuk membuat lamaran pekerjaan.

Berikut adalah data CV pelamar yang harus kamu gunakan sebagai referensi untuk membuat draft email lamaran:

=== Data CV Pelamar ===
%s
======================

Buat email lamaran yang disesuaikan dengan deskripsi pekerjaan pada gambar. Hubungkan keahlian dan pengalaman pelamar dari CV di atas dengan kebutuhan lowongan tersebut secara relevan dan logis, meskipun bidangnya berbeda. Misalnya, keahlian visual/desain bisa dihubungkan dengan kebutuhan promosi, branding, atau media sosial perusahaan.

Draf email WAJIB mengikuti format template berikut ini secara presisi, dengan mengganti semua placeholder bertanda [] menggunakan informasi yang sesuai dari data CV dan gambar lowongan:

Dengan hormat,

Sehubungan dengan lowongan posisi [Nama Posisi] yang saya temukan, saya tertarik untuk mengajukan lamaran. Dengan latar belakang [Pendidikan/Pengalaman Terakhir Anda] dan pengalaman selama [Jumlah waktu/bulan/tahun] di bidang [Bidang/Keahlian], saya pernah [Capaian/Pekerjaan 1 dari CV] serta [Capaian/Pekerjaan 2 dari CV].

Keterampilan saya dalam [Keterampilan 1], [Keterampilan 2], dan [Keterampilan 3] sangat relevan dengan persyaratan yang tercantum dalam deskripsi pekerjaan. Saya yakin dapat memberikan kontribusi yang signifikan dalam [Area spesifik pekerjaan/Tujuan lowongan].

Terlampir saya sertakan CV dan portofolio yang lebih lengkap untuk dipertimbangkan. Saya sangat berminat untuk berdiskusi lebih lanjut mengenai peluang ini.

Terima kasih atas perhatiannya.

Hormat saya,
[Nama Lengkap Pelamar]
[Alamat Email Pelamar]
No. HP/WhatsApp: 088232248005
Portofolio: https://ramadhani-portofolio.vercel.app/

Ketentuan Tambahan:
- Subjek email (subject) wajib menggunakan subjek khusus yang diminta/tertulis di dalam gambar lowongan (misalnya "Kitchen_peachy"). Jika tidak ada instruksi subjek khusus di dalam gambar, gunakan format: "Berkas Lamaran - [Nama Lengkap Pelamar] - [Nama Posisi]"
- Draf email wajib menggunakan baris kosong (double newline) untuk memisahkan setiap paragraf agar terlihat rapi dan terstruktur.
- Gunakan gaya bahasa yang simpel tetapi tetap profesional dan pastikan juga agar penggunaan bahasanya seperti asli manusia jangan seperti bahasa ai.
- Jangan menggunakan kata-kata overclaim atau berlebihan seperti "profesional", "professional", "ahli", "expert", dll. Tuliskan peran atau latar belakang secara jujur dan apa adanya sesuai dengan riwayat pada CV (misalnya: "lulusan Komunikasi Visual", "pernah bekerja sebagai video editor freelance", "memiliki pengalaman membuat konten media sosial").
- Jangan sertakan placeholder kosong atau teks tanda kurung siku [] dalam draf akhir. Semua harus terisi menggunakan data nyata dari CV.`, cvSummary)

	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{
			Data:     imageBytes,
			MIMEType: mimeType,
		}},
	}

	var resp *genai.GenerateContentResponse
	var lastErr error

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, lastErr = s.client.Models.GenerateContent(
			ctx,
			s.model,
			[]*genai.Content{{Parts: parts}},
			config,
		)
		if lastErr == nil {
			break
		}

		if attempt < maxAttempts {
			delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			fmt.Printf("API request failed (attempt %d/%d). Retrying in %v: %v\n", attempt, maxAttempts, delay, lastErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to generate content after %d attempts: %w", maxAttempts, lastErr)
	}

	rawText := strings.TrimSpace(resp.Text())
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var details JobDetails
	if err := json.Unmarshal([]byte(rawText), &details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w. Response text: %s", err, resp.Text())
	}

	return &details, nil
}

func detectMIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}
