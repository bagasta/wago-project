# WAGO Deployment Maintenance Guide

Panduan rutin untuk merawat layanan di VPS `wago-wms.chiefaiofficer.id`. Semua perintah diasumsikan dijalankan sebagai `root` di `/root/wago-project`.

## 1. Struktur dan Lokasi Penting
- Backend binary + config: `/root/wago-project/backend/wago`, `.env` di folder yang sama.
- WhatsApp store: `/root/wago-project/backend/whatsapp-sessions` (backup sebelum migrasi besar).
- Frontend build output: `/root/wago-project/frontend/dist`.
- Nginx site config: `/etc/nginx/sites-available/wago-wms.conf` (symlink ke `sites-enabled/`).
- TLS files: `/etc/letsencrypt/live/wago-wms.chiefaiofficer.id/`.
- DB target: `postgresql://postgres:aiagronomists@194.238.23.242:5432/wago`.

## 2. Backend Service (systemd `wago`)
1. **Status & Logs**
   ```bash
   systemctl status wago
   journalctl -u wago -f
   ```
2. **Start/Stop/Restart**
   ```bash
   systemctl restart wago
   systemctl stop wago
   systemctl start wago
   ```
3. **Konfigurasi**
   - Edit `/root/wago-project/backend/.env` untuk mengganti `APP_PORT`, `DATABASE_URL`, dsb.
   - Setelah mengubah `.env`, jalankan `systemctl restart wago`.
   - Binary saat ini membutuhkan jalur migrasi lama `/home/bagas/wago-project/backend/migrations`. Pastikan symlink ke `/root/wago-project/backend/migrations` tetap ada:
     ```bash
     ln -sf /root/wago-project/backend/migrations /home/bagas/wago-project/backend/migrations
     ```
4. **Re-build Backend**
   - Bila ingin membangun ulang dari source:
     ```bash
     cd /root/wago-project/backend
     env GOPATH=/root/wago-project/.gopath \
         GOMODCACHE=/root/wago-project/.gomodcache \
         GOCACHE=/root/wago-project/.gocache-build \
         go build -o wago cmd/server/main.go
     systemctl restart wago
     ```
   - Pastikan Go 1.24+ tersedia (gunakan toolchain download bawaan Go bila perlu).
5. **Kesehatan**
   - Endpoint internal: `/health`. Uji dari luar sandbox, mis. `curl https://wago-wms.chiefaiofficer.id/api/v1/health` jika sudah diproksi.
   - Jika port bentrok, ubah `APP_PORT` lalu perbarui proxy di Nginx.

## 3. Database PostgreSQL
- Server: `194.238.23.242`, user `postgres`, DB `wago`.
- Tes koneksi:
  ```bash
  psql postgresql://postgres:aiagronomists@194.238.23.242:5432/postgres -c '\l'
  ```
- Migrasi dijalankan otomatis saat service naik. Lihat tabel `schema_migrations` bila perlu audit.
- Backup rutin: gunakan `pg_dump postgresql://.../wago > wago-$(date +%F).sql`.

## 4. Frontend Maintenance
1. **Dependencies & Build**
   ```bash
   cd /root/wago-project/frontend
   npm ci
   npm run build
   ```
   > Catatan: Vite 7 secara resmi membutuhkan Node 20.19+; gunakan Node 20 LTS saat build untuk menghindari peringatan.
2. **Hasil build** otomatis disajikan oleh Nginx dari `frontend/dist`. Tidak perlu restart service.

## 5. Nginx & TLS
1. **Konfigurasi**
   - File utama: `/etc/nginx/sites-available/wago-wms.conf`.
   - Backend proxy diarahkan ke `127.0.0.1:8081` (`/api/` & `/ws/`).
   - SPA served dari `root /root/wago-project/frontend/dist`.
2. **Reload**
   ```bash
   nginx -t
   systemctl reload nginx
   ```
3. **TLS (Certbot)**
   - Sertifikat otomatis dibuat oleh `certbot --nginx`.
   - Cek status timer: `systemctl list-timers | grep certbot`.
   - Tes pembaruan: `certbot renew --dry-run`.
   - File sertifikat:
     - `fullchain.pem`: `/etc/letsencrypt/live/wago-wms.chiefaiofficer.id/fullchain.pem`
     - `privkey.pem`: `/etc/letsencrypt/live/wago-wms.chiefaiofficer.id/privkey.pem`

## 6. Monitoring & Troubleshooting
- **Backend tidak mau start**
  - Periksa `journalctl -u wago` untuk error koneksi DB / port.
  - Pastikan database `wago` tersedia dan kredensial benar.
  - Pastikan port pada `.env` tidak bentrok (cek `ss -tulpn | grep <port>`).
- **Frontend tidak update**
  - Pastikan `npm run build` selesai tanpa error.
  - Bersihkan cache browser setelah deploy.
- **Nginx 502/504**
  - Cek apakah `wago.service` aktif.
  - Periksa `tail -f /var/log/nginx/error.log`.
- **Sertifikat kadaluarsa**
  - Jalankan `certbot renew`.
  - Reload Nginx setelah sertifikat diperbarui.
- **Backup rutin**
  - Simpan salinan `.env`, directory `whatsapp-sessions`, dan dump DB secara berkala.

## 7. Checklist Pasca Update
1. Pull perubahan git atau salin file terbaru.
2. Update `.env` bila ada variabel baru.
3. Rebuild backend (jika ada perubahan Go).
4. `npm ci && npm run build` untuk frontend.
5. `nginx -t && systemctl reload nginx`.
6. `systemctl restart wago` dan verifikasi `status`.
7. Tes `https://wago-wms.chiefaiofficer.id` (login, QR scan, API, WebSocket).
