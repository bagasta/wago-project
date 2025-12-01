# Frontend - WhatsApp Multi-Session Dashboard

Panduan lengkap untuk menjalankan dan menggunakan Frontend dashboard.

## 1. Prasyarat (Prerequisites)

Pastikan Anda telah menginstal:
- **Node.js** versi 18 atau lebih baru.
- **NPM** (biasanya terinstal bersama Node.js).

## 2. Instalasi

Masuk ke folder `frontend` dan instal dependensi:

```bash
cd frontend
npm install
```

## 3. Konfigurasi

Frontend dikonfigurasi untuk terhubung ke Backend melalui proxy Vite.
Secara default, proxy mengarah ke `http://localhost:8080`.

Jika Anda perlu mengubah target backend, edit file `vite.config.js`:
```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080', // Ubah sesuai URL backend
      changeOrigin: true,
    },
    '/ws': {
      target: 'ws://localhost:8080', // Ubah sesuai URL backend
      ws: true,
    }
  }
}
```

## 4. Menjalankan Development Server

Untuk menjalankan aplikasi dalam mode development:

```bash
npm run dev
```
Atau jika ingin bisa diakses dari network lain (misal HP):
```bash
npm run dev -- --host
```

Aplikasi akan berjalan di `http://localhost:5173`.

## 5. Build untuk Production

Untuk membuat build production yang optimal:

```bash
npm run build
```

Hasil build akan berada di folder `dist/`. Anda bisa menyajikan folder ini menggunakan web server statis (Nginx, Apache, atau `serve`).

## 6. Cara Penggunaan Dashboard

1.  **Generate PIN**:
    - Buka halaman utama, klik link "Generate one here".
    - Klik tombol "Generate Secure PIN".
    - **PENTING**: Simpan PIN yang muncul. PIN ini tidak akan ditampilkan lagi.

2.  **Login**:
    - Masukkan 6-digit PIN yang telah Anda generate.
    - Klik "Access Dashboard".

3.  **Membuat Sesi**:
    - Di Dashboard, klik tombol "+ Add Session".
    - Masukkan Nama Sesi (contoh: "Admin Bot") dan Webhook URL (opsional).
    - Klik "Create Session".

4.  **Menghubungkan WhatsApp**:
    - Pada kartu sesi yang baru dibuat, klik tombol "Connect".
    - Akan muncul modal berisi QR Code.
    - Buka WhatsApp di HP Anda -> Menu (titik tiga) -> Linked Devices -> Link a Device.
    - Scan QR Code yang muncul di layar.
    - Tunggu hingga status berubah menjadi hijau (**CONNECTED**).

5.  **Menghapus Sesi**:
    - Klik tombol ikon tempat sampah (Delete) pada kartu sesi yang ingin dihapus.
