BEGIN;

INSERT INTO permissions (name, description) VALUES
    ('student:list', 'Melihat daftar seluruh mahasiswa'),
    ('student:read:any', 'Melihat data mahasiswa mana pun'),
    ('student:create', 'Membuat data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete', 'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT (role_name, permission_name) DO NOTHING;

ALTER TABLE students
    ADD COLUMN IF NOT EXISTS owner_id INTEGER;

UPDATE students
SET owner_id = NULL
WHERE owner_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM users
      WHERE users.id = students.owner_id
  );

ALTER TABLE students
    DROP CONSTRAINT IF EXISTS students_owner_id_fkey;

ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS students_owner_id_idx
    ON students (owner_id);

COMMIT;
