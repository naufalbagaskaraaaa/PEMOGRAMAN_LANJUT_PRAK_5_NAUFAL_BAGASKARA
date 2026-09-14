# API Students

REST API mahasiswa dengan Fiber, PostgreSQL, dan Repository Pattern.

## Skema Tabel

Migrasi ada di `migrations/001_create_students.sql`. Autentikasi memakai `migrations/003_auth.sql`, yang membuat tabel `users` bila belum tersedia dan tabel `refresh_tokens`.

| Kolom | Tipe | Batasan |
| --- | --- | --- |
| `id` | `SERIAL` | Primary key |
| `nim` | `VARCHAR(32)` | Wajib, unik |
| `name` | `VARCHAR(120)` | Wajib |
| `grade` | `NUMERIC(3,2)` | Wajib, 0 sampai 4 |
| `is_active` | `BOOLEAN` | Wajib, default `TRUE` |
| `created_at` | `TIMESTAMPTZ` | Wajib, default waktu saat insert |

Indeks `idx_students_name` membantu pencarian dan pengurutan berdasarkan nama. Keunikan NIM dijaga oleh database karena constraint berlaku konsisten untuk semua client dan aman terhadap race condition saat dua request berjalan bersamaan. Kode Go tetap menerjemahkan pelanggaran constraint menjadi `409 Conflict`.

## Menyiapkan Database

1. Buat database PostgreSQL, misalnya `latihan_fiber`.
2. Salin `.env.example` menjadi `.env`, lalu isi kredensial PostgreSQL.
3. Jalankan `migrations/001_create_students.sql` pada database tersebut, misalnya dengan `psql`.
4. Jalankan migrasi lain sesuai urutan: `002_create_prestasi.sql`, lalu `003_auth.sql`.
5. Isi `JWT_SECRET` dengan secret acak yang panjang, lalu jalankan API dengan `go run .`.

Untuk database praktikum yang sudah berisi akun lama, kosongkan data akun sebelum migrasi autentikasi dengan `TRUNCATE TABLE users CASCADE;`. Langkah ini hanya untuk database praktikum karena bersifat destruktif. Password lama tidak dapat dipulihkan atau di-hash ulang tanpa mengetahui plaintext password; karena itu akun harus dibuat ulang melalui endpoint register.

Endpoint health: `GET /health`. Endpoint ini mengembalikan `503 Service Unavailable` jika PostgreSQL tidak dapat di-ping karena request tidak dapat dilayani tanpa database.

## Environment Variables

`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, dan `JWT_SECRET` dibaca dari `.env`. Berkas `.env` diabaikan Git; gunakan `.env.example` sebagai template. `JWT_SECRET` wajib diisi dan tidak memiliki fallback.

Password disimpan menggunakan bcrypt cost `12` (di atas batas minimal 10). Cost ini memperlambat brute-force secara signifikan dibanding cost rendah, tetapi tetap realistis untuk waktu respons API pada perangkat praktikum. Password tidak pernah dikirim dalam respons karena field `PasswordHash` memakai JSON tag `-`.

Middleware autentikasi menolak algoritma JWT selain HS256, termasuk `alg: none`. Token yang kedaluwarsa menghasilkan `401` dengan pesan `token kedaluwarsa`, sedangkan token rusak atau signature salah menghasilkan `401` dengan pesan `token tidak valid`; pembedaan ini aman karena tidak membocorkan kredensial login. Sebaliknya, login selalu memakai pesan `username atau password salah` untuk username tidak ditemukan maupun password salah agar user enumeration lebih sulit. Setelah lima kegagalan login berturut-turut per username dan alamat IP, percobaan keenam dibatasi dengan `429` dan header `Retry-After` selama lima menit.

Untuk laporan, buktikan dengan `go test ./...` dan tangkapan layar database bahwa `users.password_hash` berisi string bcrypt, bukan password plaintext. Tangkapan layar pengujian HTTP perlu memperlihatkan status response dan header, terutama `WWW-Authenticate`, `Retry-After`, serta status `401` untuk refresh token yang sudah dirotasi.

## API Contract

| Metode | Endpoint | Parameter / Query | Contoh body | Kemungkinan status | Contoh respons |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active` | - | `200`, `401`, `503` | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[],"meta":{"page":1,"limit":10,"total":0,"total_pages":0}}` |
| `GET` | `/api/v1/students/:id` | `id` | - | `200`, `400`, `401`, `404` | `{"success":true,"message":"mahasiswa ditemukan","data":{...}}` |
| `POST` | `/api/v1/students` | - | `{"nim":"A01","name":"Budi","grade":3.5}` | `201`, `401`, `409`, `415`, `422` | `{"success":true,"message":"mahasiswa berhasil dibuat","data":{...}}` |
| `PUT` | `/api/v1/students/:id` | `id` | `{"nim":"A01","name":"Budi","grade":3.5,"is_active":true}` | `200`, `400`, `401`, `404`, `409`, `415`, `422` | `{"success":true,"message":"data mahasiswa berhasil diganti seluruhnya","data":{...}}` |
| `PATCH` | `/api/v1/students/:id` | `id` | `{"name":"Budi Baru"}` | `200`, `400`, `401`, `404`, `409`, `415`, `422` | `{"success":true,"message":"data mahasiswa berhasil diperbarui sebagian","data":{...}}` |
| `DELETE` | `/api/v1/students/:id` | `id` | - | `204`, `401`, `404` | tanpa body |
| `POST` | `/api/v1/auth/register` | - | `{"username":"budi","email":"budi@example.com","password":"Rahasia123"}` | `201`, `415`, `409`, `422` | `{"success":true,"message":"pengguna berhasil didaftarkan","data":{...}}` |
| `POST` | `/api/v1/auth/login` | - | `{"username":"budi","password":"Rahasia123"}` | `200`, `401`, `415`, `422` | `{"success":true,"message":"berhasil masuk","data":{"access_token":"...","refresh_token":"..."}}` |

Pencarian nama memakai `ILIKE`, filter menggunakan parameter query, pengurutan dibatasi whitelist kolom, dan paginasi memakai `LIMIT` serta `OFFSET`. Nilai `meta.total` dihitung dengan `SELECT COUNT(*)` menggunakan filter yang sama.

## Peta Arsitektur

| Folder/package | Layer | Tanggung jawab |
| --- | --- | --- |
| `app/model` | Domain | Entity, request model, dan response model tanpa dependency package internal lain |
| `app/service` | Application | Business rules dan service/controller yang mengatur alur request |
| `app/repository` | Infrastructure | Interface repository dan query PostgreSQL |
| `database` | Infrastructure | Pembuatan connection pool dan ping database |
| `config` | Infrastructure | Environment dan konfigurasi logger Lumberjack |
| `helper` | Interface adapter | Parsing request dan response HTTP umum |
| `middleware` | Interface adapter | Request ID, logger HTTP, CORS, dan error handler |
| `route` | Interface adapter | Pendaftaran alamat endpoint tanpa SQL atau validasi bisnis |
| `main.go` | Composition root | Merakit dependency dan menjalankan aplikasi |

```text
main.go -> middleware, route -> app/service -> app/repository -> database/PostgreSQL
						 \-> helper
app/model menjadi dependency data bersama tanpa bergantung pada Fiber atau database.
```

Struktur ini disederhanakan dibandingkan Clean Architecture murni karena belum memisahkan use case, port, adapter, dan dependency injection container menjadi package terpisah. Service masih menerima `fiber.Ctx` agar controller HTTP tetap ringkas. Konsekuensinya, pengujian business rules tetap murni, tetapi pengujian service/controller masih bergantung pada Fiber dan mock repository.

## Pemeriksaan Kebocoran Layer

- `app/model` hanya memakai standard library (`time`) dan tidak mengimpor package internal project.
- `app/repository` tidak mengimpor Fiber; package ini hanya bergantung pada model dan pgx.
- `app/service` tidak berisi SQL; query hanya berada di repository.
- `route/route.go` hanya mendaftarkan routes dan middleware.
- `main.go` hanya memuat konfigurasi, pool, repository, service, middleware, routes, dan server.

Saat restrukturisasi, logika validasi dan penerapan PATCH yang sebelumnya berada di handler dipindahkan ke `app/service/student_rules.go`. Pendaftaran routes dipindahkan dari `main.go` ke `route/route.go`, sedangkan health check dan error handler dipindahkan ke service/middleware.

## Laporan

Laporan dikumpulkan sebagai `Tugas4_NIM.pdf` dan memuat tautan repositori GitHub pada halaman pertama, struktur folder akhir, peta arsitektur, diagram dependency, potongan kode penting, output `go test ./...`, screenshot endpoint, serta penjelasan bantuan yang digunakan termasuk tools AI dan bagian yang dibantu.