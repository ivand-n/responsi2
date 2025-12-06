# Responsi Market (Go Backend + Flutter Frontend)

Ivan Darmawan
H1D022042
SHIFT B - I

## Prasyarat
- Go 1.22+
- Flutter 3.x (Dart SDK terikut)
- PostgreSQL Supabase (project & API keys)
- Akses IPv6 (Supabase free tier hanya IPv6) atau VPN dengan IPv6 (mis. WARP)
- Git, Node optional (untuk tools linting)

## Struktur
- `backend/` — API Go native (JWT auth, CRUD barang)
- `cmarket/` — Aplikasi Flutter (Google Sign-In, panggil backend)

## Konfigurasi Environment
Buat/isi `backend/.env`:
```
DB_URL=postgresql://postgres:<PASSWORD>@db.<PROJECT-REF>.supabase.co:5432/postgres?sslmode=require
SUPABASE_JWT_SECRET=<JWT_SECRET_SUPABASE>
```
Catatan: gunakan `sslmode=require` untuk Supabase.

## Menjalankan Backend (Go)
```bash
cd backend
go mod tidy
go run .
```
Server jalan di `http://localhost:8080`.

## Endpoint Utama (dengan Bearer JWT)
- `GET /api/barang` — list barang
- `POST /api/barang` — tambah
- `PUT /api/barang/{id}` — update milik user
- `DELETE /api/barang/{id}` — hapus milik user

## Menjalankan Frontend (Flutter)
```bash
cd cmarket
flutter pub get
flutter run
```

## Konfigurasi Frontend Supabase & OAuth
Di `lib/main.dart`:
- `Supabase.initialize(url, anonKey)` → isi `url` dengan URL project Supabase (bukan connection string DB; gunakan `https://<PROJECT-REF>.supabase.co`).
- Pastikan `anonKey` sesuai dari Supabase Settings → API.
- Google OAuth: set `webClientId`/`iosClientId` sesuai kredensial OAuth yang didaftarkan.

## Menghubungkan ke Backend
- Atur `url` di `_panggilBackendGo` (HomePage) ke host backend:
  - Emulator Android: `http://10.0.2.2:8080`
  - Device fisik: `http://<IP-Laptop>:8080`
- Header `Authorization: Bearer <accessToken>` otomatis dikirim.

## Troubleshooting
- **Tidak bisa konek DB (IPv6 only)**: aktifkan IPv6 pada adaptor Windows atau gunakan VPN dengan IPv6.
- **Token invalid**: pastikan `SUPABASE_JWT_SECRET` sama dengan JWT Secret Supabase dan gunakan access token dari sesi Supabase.
- **CORS**: backend sudah mengizinkan `*`; sesuaikan bila perlu untuk produksi.
