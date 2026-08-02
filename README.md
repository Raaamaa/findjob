# Jobber Bot

Jobber adalah Bot Telegram otomatis untuk melamar pekerjaan secara instan menggunakan analisis AI (Gemini) langsung dari screenshot lowongan kerja.

## Cara Menggunakan

1. **Konfigurasi Lingkungan**
   Salin berkas `.env.example` menjadi `.env` dan lengkapi variabelnya:
   ```bash
   cp .env.example .env
   chmod 600 .env
   ```

2. **Kompilasi Program**
   ```bash
   go build -o jobber main.go
   ```

3. **Menjalankan Bot**
   * **Mode Standar** (Terminal harus tetap terbuka):
     ```bash
     go run main.go
     ```
   * **Mode Latar Belakang (Background)** (Aman saat terminal ditutup):
     ```bash
     nohup ./jobber > bot.log 2>&1 &
     ```

4. **Menghentikan Bot (Mode Latar Belakang)**
   ```bash
   pkill jobber
   ```

5. **Melihat Log Aktivitas**
   ```bash
   tail -f bot.log
   ```

## Fitur & Perintah Bot Telegram

Kirim `/start` pada bot Telegram Anda untuk memulai, lalu kirim screenshot lowongan kerja.

* **/edit <email/subject/body> <nilai_baru>**: Koreksi detail data lowongan sebelum dikirim.
* **/clear** atau **.clear**: Hapus sesi lamaran aktif saat ini.
* **Accept / Deny**: Tombol interaktif untuk menyetujui pengiriman email lamaran beserta CV PDF terlampir.

## Pengujian
Jalankan tes otomatis:
```bash
go test -v ./...
```
