CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    grade NUMERIC(3, 2) NOT NULL CHECK (grade >= 0 AND grade <= 4),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_students_name ON students (name);
