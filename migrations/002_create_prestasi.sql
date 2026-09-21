CREATE TABLE IF NOT EXISTS public.prestasi (
	id_prestasi VARCHAR(32) PRIMARY KEY,
	nama_prestasi CHAR(100) NOT NULL,
	juara INT UNIQUE,
	nim VARCHAR(32),
	CONSTRAINT fk_prestasi_mahasiswa
		FOREIGN KEY (nim)
		REFERENCES public.students(nim)
);