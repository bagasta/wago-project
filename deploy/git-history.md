# Git History & Maintenance Log

Catatan aktivitas penting yang dilakukan pada repo dan server produksi.

| Timestamp (WIB) | Aktivitas | Rincian | Referensi |
| --- | --- | --- | --- |
| 2025-12-02 14:42 | Rebuild backend pasca `git pull` | `go build -o backend/wago cmd/server/main.go` dengan cache lokal sementara, lalu `systemctl restart wago` memastikan sesi WhatsApp otomatis reconnect. | `systemctl status wago` PID 67983 |
| 2025-12-02 14:58 | Membersihkan cache Go & rebase `main` | Hapus `.gocache` dan `.gomodcache`, rebase terhadap `origin/main`, rebuild backend menggunakan cache `/tmp`, restart service. | Commit `d593c9e` |
| 2025-12-02 15:00 | Tambah `.gitignore` root dan push housekeeping | Dokumentasikan aturan ignore baru, hapus artefak build dari Git history, push `chore: ignore local build caches` & `chore: add repo-level gitignore`. | Commit `d593c9e`, `150e8b4` |
| 2025-12-02 16:00 | Validasi toggle group response | Pull pembaruan frontend, rebuild backend, restart service memastikan endpoint siap menerima toggle group. | `systemctl status wago` PID 88783 |
| 2025-12-02 16:02 | Tambah route PUT `/api/v1/sessions/{id}` | Menambah handler update session ke router agar tombol toggle mengirim request valid; rebuild & restart backend. | Commit `d9f6a18` |
