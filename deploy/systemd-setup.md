# Setup tanpa Docker (systemd)

Panduan singkat menjalankan backend lewat systemd dan menyajikan frontend statik. Sesuaikan path repo kalau tidak berada di `/home/clevio/stack/wago-project`.

## Prasyarat
- Debian/Ubuntu dengan systemd, Postgres 15+, Go 1.24+, Node.js 20+ (npm), dan Nginx untuk frontend.
- Buat database + user Postgres, contoh:
  ```bash
  sudo -u postgres psql <<'SQL'
  CREATE DATABASE wago;
  CREATE USER wago WITH ENCRYPTED PASSWORD 'ganti-password';
  GRANT ALL PRIVILEGES ON DATABASE wago TO wago;
  SQL
  ```

## 1) Siapkan env backend
```bash
REPO_DIR=/home/clevio/stack/wago-project
cd $REPO_DIR/backend
cp .env.example .env
# Isi:
# APP_PORT=8081                        # selaraskan dengan nginx
# DATABASE_URL=postgres://wago:...@localhost:5432/wago?sslmode=disable
# JWT_SECRET=acak-panjang
# WHATSAPP_DATA_DIR=/home/clevio/stack/wago-project/backend/whatsapp-sessions
# ALLOWED_ORIGINS=https://domain-frontend-anda
WHATSAPP_DATA_DIR=$(grep -E '^WHATSAPP_DATA_DIR' .env | cut -d= -f2)
mkdir -p "$WHATSAPP_DATA_DIR"
```

## 2) Build backend
```bash
cd $REPO_DIR/backend
go build -o wago cmd/server/main.go
```

## 3) Install service systemd
```bash
sudo sed "s#/root/wago-project#$REPO_DIR#g" \
  $REPO_DIR/deploy/wago.service | sudo tee /etc/systemd/system/wago.service
# Ganti User=... pada unit jika tidak ingin menjalankan sebagai root.
sudo systemctl daemon-reload
sudo systemctl enable --now wago
sudo systemctl status wago --no-pager
```
Log jalan: `sudo journalctl -u wago -f`

## 4) Build frontend
```bash
cd $REPO_DIR/frontend
npm ci
npm run build
```

## 5) Nginx
Sesuaikan `server_name`, path `root`, dan port backend (`proxy_pass`) di `deploy/nginx-wago-wms.conf`, lalu pasang:
```bash
sudo sed "s#/root/wago-project#$REPO_DIR#g" \
  $REPO_DIR/deploy/nginx-wago-wms.conf | sudo tee /etc/nginx/sites-available/wago.conf
sudo ln -sf /etc/nginx/sites-available/wago.conf /etc/nginx/sites-enabled/wago.conf
sudo nginx -t && sudo systemctl reload nginx
```

## 6) Smoke test
- `curl -f http://127.0.0.1:8081/health`
- Buka halaman login frontend (domain yang Anda set di Nginx), pastikan QR muncul dan API/WebSocket mengarah ke domain yang sama.

## 7) Routine deploy
```bash
cd $REPO_DIR
git pull
cd backend && go build -o wago cmd/server/main.go
sudo systemctl restart wago
cd ../frontend && npm run build
sudo nginx -t && sudo systemctl reload nginx
```
