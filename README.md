# 🚀 Jobber Bot - Panduan Penggunaan Lengkap

**Jobber Bot** adalah bot Telegram otomatis untuk menganalisis screenshot lowongan kerja menggunakan AI (Gemini Vision), menyusun draf surat lamaran (*cover letter*) secara personal berdasarkan CV Anda, dan mengirimkan email lamaran beserta lampiran CV PDF secara langsung.

---

## 📌 1. Persiapan & Setup Awal

Sebelum menjalankan bot, pastikan Anda telah menyiapkan konfigurasi berikut pada file **[`.env`](file://.env)**:

### A. Mendapatkan Gemini API Key
1. Buka [Google AI Studio](https://aistudio.google.com/app/apikey).
2. Login dengan akun Google Anda.
3. Klik **Create API Key**, lalu salin kunci API tersebut.
4. Masukkan ke file `.env`:
   ```ini
   GEMINI_API_KEY=AIzaSy...
   GEMINI_MODEL=gemini-2.5-pro
   ```

### B. Membuat Bot Telegram & Token
1. Buka aplikasi Telegram, cari akun bot resmi **[@BotFather](https://t.me/BotFather)**.
2. Kirim perintah `/newbot` dan ikuti petunjuknya hingga selesai.
3. Salin **HTTP API Token** yang berformat seperti `123456789:ABCdef...`.
4. Masukkan token tersebut dan username Telegram Anda ke file `.env`:
   ```ini
   TELEGRAM_BOT_TOKEN=123456789:ABCdef...
   ALLOWED_TELEGRAM_USERNAME=username_telegram_anda
   ```

### C. Konfigurasi SMTP Gmail (Kunci Sandi Aplikasi / App Password)
1. Aktifkan **Verifikasi 2 Langkah (2-Step Verification)** di [Keamanan Akun Google](https://myaccount.google.com/security).
2. Cari menu **Sandi Aplikasi (App Passwords)** di pencarian akun Google.
3. Buat Sandi Aplikasi baru (misal diberi nama `Jobber Bot`), lalu salin **16 karakter kode sandi** yang dihasilkan.
4. Masukkan ke file `.env`:
   ```ini
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=email_anda@gmail.com
   SMTP_PASS=kodesandi16digit
   DEFAULT_SENDER_NAME=Akhmad Rizki Ramadhani
   ```

### D. File CV PDF & Ringkasan CV
Pastikan dua file ini berada di dalam direktori bot:
- `CV Akhmad Rizki Ramadhani ( English ).pdf` (File PDF CV fisik yang dilampirkan ke email).
- `cv.md` (Ringkasan CV teks yang dibaca oleh Gemini AI).

---

## 🏃 2. Cara Menjalankan Bot

### A. Mode Pengujian (Terminal Langsung)
```bash
go run main.go
```

### B. Mode Produksi (Kompilasi & Latar Belakang / Background)
Jika Anda ingin bot tetap berjalan setelah terminal ditutup:

1. **Kompilasi Program**:
   ```bash
   go build -o jobber main.go
   ```
2. **Jalankan di Background**:
   ```bash
   nohup ./jobber > bot.log 2>&1 &
   ```
3. **Melihat Log Aktivitas**:
   ```bash
   tail -f bot.log
   ```
4. **Menghentikan Bot**:
   ```bash
   pkill jobber
   ```

---

## 💬 3. Panduan Penggunaan di Telegram

### Langkah 1: Memulai Bot
Buka chat bot Anda di Telegram dan kirim perintah `/start`.

### Langkah 2: Mengirim Screenshot Lowongan Kerja
Kirimkan foto atau screenshot gambar lowongan pekerjaan ke dalam chat Telegram bot.

### Langkah 3: Menerima Hasil Analisis AI
Bot akan merespon secara bertahap:
1. Menerima gambar & mengunduh ke memori.
2. Membaca data CV dan menganalisis lowongan via Gemini AI.
3. Menampilkan **Preview Detail Lowongan & Draf Email Lamaran** lengkap dengan tombol interaktif **Accept (Kirim)** dan **Deny (Batal)**.

### Langkah 4: Mengedit Data (Opsional)
Jika terdapat kesalahan ekstraksi (misalnya email tidak terdeteksi di gambar atau nama posisi kurang pas), Anda dapat mengoreksinya sebelum dikirim:
- `/edit company <Nama Perusahaan Baru>`
- `/edit position <Nama Posisi Baru>`
- `/edit email <alamat_email_tujuan>`
- `/edit subject <Subjek Email Baru>`
- `/edit body <Isi Draf Email Baru>`

Setelah menjalankan perintah `/edit`, bot akan memperbarui dan menampilkan preview terbaru beserta tombol konfirmasinya.

### Langkah 5: Mengirim Lamaran
- Klik tombol **Accept (Kirim)**: Bot akan mengirimkan email lamaran + lampiran PDF CV Anda secara langsung ke alamat email tujuan, mencatat riwayat lamaran ke `history.json`, dan mengosongkan sesi.
- Klik tombol **Deny (Batal)**: Membatalkan pengiriman email.
- Perintah `/clear` atau `.clear`: Menghapus sesi aktif tanpa mengirim apapun jika Anda ingin mengulang dari awal.

---

## ❓ 4. Peringatan & Fitur Penting

1. **Peringatan Email Tidak Terdeteksi**:
   Jika gambar lowongan kerja tidak memuat alamat email, bot akan menampilkan peringatan:
   `⚠️ PERINGATAN: Email penerima tidak terdeteksi...`
   Gunakan perintah `/edit email hrd@perusahaan.com` untuk memasukkan alamat email tujuan sebelum menekan Accept.

2. **Deteksi Duplikasi Lamaran**:
   Bot secara otomatis mendeteksi jika:
   - Gambar lowongan yang sama dikirim lebih dari sekali.
   - Posisi dan email penerima yang sama pernah Anda lamar sebelumnya.

3. **Ganti Model AI**:
   Ubah `GEMINI_MODEL` pada file `.env` ke `gemini-2.5-pro` untuk kualitas teks tertinggi, atau `gemini-2.5-flash` untuk kecepatan ekstraksi tercepat.
