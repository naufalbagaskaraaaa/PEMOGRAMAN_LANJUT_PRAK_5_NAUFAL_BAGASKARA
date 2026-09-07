# API Students

REST API mahasiswa dengan Fiber, PostgreSQL, dan Repository Pattern.

## Skema Tabel

Migrasi ada di `migrations/001_create_students.sql`.

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
4. Jalankan API dengan `go run .`.

Endpoint health: `GET /health`. Endpoint ini mengembalikan `503 Service Unavailable` jika PostgreSQL tidak dapat di-ping karena request tidak dapat dilayani tanpa database.

## Environment Variables

`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, dan `DB_NAME` dibaca dari `.env`. Berkas `.env` diabaikan Git; gunakan `.env.example` sebagai template.

## API Contract

| Metode | Endpoint | Parameter / Query | Body Request | Status |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active` | - | `200` |
| `GET` | `/api/v1/students/:id` | `id` | - | `200`, `400`, `404` |
| `POST` | `/api/v1/students` | - | `nim`, `name`, `grade` | `201`, `409`, `422` |
| `PUT` | `/api/v1/students/:id` | `id` | `nim`, `name`, `grade`, `is_active` | `200`, `404`, `409`, `422` |
| `PATCH` | `/api/v1/students/:id` | `id` | field yang diubah | `200`, `400`, `404`, `409`, `422` |
| `DELETE` | `/api/v1/students/:id` | `id` | - | `204`, `404` |

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
- `route/route.go` hanya mendaftarkan route dan middleware.
- `main.go` hanya memuat konfigurasi, pool, repository, service, middleware, route, dan server.

Saat restrukturisasi, logika validasi dan penerapan PATCH yang sebelumnya berada di handler dipindahkan ke `app/service/student_rules.go`. Pendaftaran route dipindahkan dari `main.go` ke `route/route.go`, sedangkan health check dan error handler dipindahkan ke service/middleware.

## Laporan

Laporan dikumpulkan sebagai `Tugas4_NIM.pdf` dan memuat tautan repositori GitHub pada halaman pertama, struktur folder akhir, peta arsitektur, diagram dependency, potongan kode penting, output `go test ./...`, screenshot endpoint, serta penjelasan bantuan yang digunakan termasuk tools AI dan bagian yang dibantu.