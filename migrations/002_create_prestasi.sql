create table prestasi (
id_prestasi varchar(32) primary key,
nama_prestasi char(100) not null,
juara int unique,
nim varchar(32),
constraint fk_prestasi_mahasiswa
foreign key (nim)
references students(nim)
);

# mantap