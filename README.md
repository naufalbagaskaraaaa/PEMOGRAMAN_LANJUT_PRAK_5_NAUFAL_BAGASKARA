# API Documentation - Student Management System

## API Contract

| Metode | Endpoint | Parameter / Query | Body Request | Status | Respon |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active` | - | `200 OK` | `{"success": true, "data": [...], "meta": {...}}` |
| **GET** | `/api/v1/students/:id` | `id` (path parameter) | - | `200 OK`<br>`400 Bad Request`<br>`404 Not Found` | Data mahasiswa / Pesan error |
| **POST** | `/api/v1/students` | - | `{"nim": "112233", "name": "Budi", "grade": 3.75}` | `201 Created`<br>`409 Conflict`<br>`415 Unsupported`<br>`422 Unprocessable` | Data mahasiswa baru + Header `Location` |
| **PUT** | `/api/v1/students/:id` | `id` (path parameter) | `{"nim": "112233", "name": "Budi A", "grade": 3.8, "is_active": true}` | `200 OK`<br>`409 Conflict`<br>`422 Unprocessable` | Data hasil update total |
| **PATCH** | `/api/v1/students/:id` | `id` (path parameter) | `{"grade": 3.9}` | `200 OK`<br>`422 Unprocessable` | Data hasil update parsial |
| **DELETE**| `/api/v1/students/:id` | `id` (path parameter) | - | `204 No Content`<br>`404 Not Found` | Tanpa body |