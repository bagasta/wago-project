# Backend - WhatsApp Multi-Session API Gateway

Panduan lengkap untuk menjalankan dan menggunakan Backend service.

## 1. Prasyarat (Prerequisites)

Pastikan Anda telah menginstal:
- **Go** (Golang) versi 1.21 atau lebih baru.
- **PostgreSQL** versi 15 atau lebih baru.
- **Make** (opsional, untuk kemudahan command).

## 2. Konfigurasi

1.  Salin file `.env.example` menjadi `.env`:
    ```bash
    cp .env.example .env
    ```

2.  Sesuaikan konfigurasi di dalam `.env`:
    ```env
    APP_PORT=8080
    DATABASE_URL=postgres://user:password@localhost:5432/wago?sslmode=disable
    JWT_SECRET=rahasia-super-aman-ganti-ini
    WHATSAPP_DATA_DIR=whatsapp-sessions
    ```
    *Catatan: Jika menggunakan socket PostgreSQL (default di beberapa Linux), `DATABASE_URL` mungkin seperti `postgres://user@/wago?host=/var/run/postgresql`.*

## 3. Setup Database

1.  Buat database baru di PostgreSQL bernama `wago` (atau sesuai konfigurasi Anda):
    ```bash
    createdb wago
    ```

2.  Migrasi database akan berjalan otomatis saat aplikasi dijalankan pertama kali.

## 4. Menjalankan Aplikasi

### Menggunakan Make (Recommended)
Dari root folder project (`wago-project/`), jalankan:
```bash
make run
```

### Menggunakan Go Command
Dari folder `backend/`, jalankan:
```bash
go run cmd/server/main.go
```

Server akan berjalan di `http://localhost:8080`.

## 5. Struktur API

### Authentication
- **POST** `/api/v1/auth/generate-pin`: Membuat PIN baru untuk user.
- **POST** `/api/v1/auth/login`: Login menggunakan PIN (Basic Auth).
- **POST** `/api/v1/auth/logout`: Logout user.

### Session Management
*(Memerlukan Header `Authorization: Bearer <token>`)*
- **POST** `/api/v1/sessions`: Membuat sesi WhatsApp baru.
- **GET** `/api/v1/sessions`: Mendapatkan daftar sesi.
- **POST** `/api/v1/sessions/{id}/start`: Memulai sesi (mendapatkan QR Code).
- **DELETE** `/api/v1/sessions/{id}`: Menghapus sesi.

### WebSocket
- **WS** `/ws/sessions/{id}?token=<token>`: Mendapatkan update realtime (QR Code, Status).

## 6. Troubleshooting

- **Database Connection Failed**: Periksa `DATABASE_URL` di `.env`. Pastikan user/password benar dan PostgreSQL service berjalan.
- **Address already in use**: Port 8080 sedang digunakan. Matikan proses yang menggunakan port tersebut atau ubah `APP_PORT` di `.env`.
