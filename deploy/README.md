# WAGO VPS Deployment (wago-wms.chiefaiofficer.id)

Panduan singkat menjalankan backend + frontend di satu VPS (user `root`, repo di `/root/wago-project`).

## Prasyarat
- Go 1.24+, Node 18+, npm, Nginx, systemd, Postgres 15+.
- Domain: `wago-wms.chiefaiofficer.id` menunjuk ke server. Sertifikat TLS bisa pakai certbot.

## 1) Siapkan env backend
```bash
cd /root/wago-project/backend
cp .env.example .env
# Set nilai berikut:
# APP_PORT=8080
# DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/wago?sslmode=disable
# JWT_SECRET=... (acak panjang)
# WHATSAPP_DATA_DIR=/root/wago-project/backend/whatsapp-sessions
# ALLOWED_ORIGINS=https://wago-wms.chiefaiofficer.id
```

## 2) Build backend binary
```bash
cd /root/wago-project/backend
go build -o wago cmd/server/main.go
```
Binary yang dipakai systemd: `/root/wago-project/backend/wago`.

## 3) Pasang service systemd
```bash
cp /root/wago-project/deploy/wago.service /etc/systemd/system/wago.service
systemctl daemon-reload
systemctl enable --now wago
systemctl status wago
```

## 4) Build frontend
```bash
cd /root/wago-project/frontend
npm ci
npm run build    # output ke dist/
```

## 5) Nginx reverse proxy + static
```bash
cp /root/wago-project/deploy/nginx-wago-wms.conf /etc/nginx/sites-available/wago-wms.conf
ln -s /etc/nginx/sites-available/wago-wms.conf /etc/nginx/sites-enabled/wago-wms.conf
nginx -t && systemctl reload nginx
```
Konfigurasi ini:
- Serve frontend statis dari `/root/wago-project/frontend/dist`
- Proxy `/api/*` dan `/ws/*` ke backend di `http://127.0.0.1:8080`

## 6) TLS (opsional, disarankan)
```bash
apt install -y certbot python3-certbot-nginx
certbot --nginx -d wago-wms.chiefaiofficer.id
systemctl reload nginx
```

## 7) Cek
- `systemctl status wago` harus `active (running)`.
- Buka `https://wago-wms.chiefaiofficer.id/login`, cek QR flow dan panggilan API/WebSocket ke domain yang sama.
