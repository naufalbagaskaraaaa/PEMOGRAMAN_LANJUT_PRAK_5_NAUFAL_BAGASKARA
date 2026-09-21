package repository

import (
	"context"
	"errors"
	"strings"

	"latihan-fiber/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPrestasiNotFound = errors.New("prestasi tidak ditemukan")

type PrestasiRepository interface {
	FindByID(ctx context.Context, idPrestasi string) (model.Prestasi, error)
	Create(ctx context.Context, req model.CreatePrestasiRequest) (model.Prestasi, error)
	FindByNIM(ctx context.Context, nim string) ([]model.Prestasi, error)
}

type prestasiPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPrestasiRepository(pool *pgxpool.Pool) PrestasiRepository {
	return &prestasiPostgresRepository{pool: pool}
}

func (r *prestasiPostgresRepository) Create(
	ctx context.Context,
	req model.CreatePrestasiRequest,
) (model.Prestasi, error) {
	var prestasi model.Prestasi
	err := r.pool.QueryRow(ctx, `
		INSERT INTO prestasi (id_prestasi, nama_prestasi, juara, nim)
		VALUES ($1, $2, $3, $4)
		RETURNING id_prestasi, nama_prestasi, juara, nim
	`, req.IDPrestasi, req.NamaPrestasi, req.Juara, req.NIM).Scan(
		&prestasi.IDPrestasi,
		&prestasi.NamaPrestasi,
		&prestasi.Juara,
		&prestasi.NIM,
	)
	if err != nil {
		return model.Prestasi{}, err
	}
	prestasi.NamaPrestasi = strings.TrimSpace(prestasi.NamaPrestasi)
	return prestasi, nil
}

func (r *prestasiPostgresRepository) FindByNIM(
	ctx context.Context,
	nim string,
) ([]model.Prestasi, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id_prestasi,
			nama_prestasi,
			juara,
			nim
		FROM prestasi
		WHERE nim = $1
		ORDER BY juara ASC
	`, nim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prestasilist := []model.Prestasi{}

	for rows.Next() {
		var prestasi model.Prestasi
		var namaPrestasi string

		err := rows.Scan(
			&prestasi.IDPrestasi,
			&namaPrestasi,
			&prestasi.Juara,
			&prestasi.NIM,
		)
		if err != nil {
			return nil, err
		}

		prestasi.NamaPrestasi = strings.TrimSpace(namaPrestasi)
		prestasilist = append(prestasilist, prestasi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prestasilist, nil
}

func (r *prestasiPostgresRepository) FindByID(
	ctx context.Context,
	idPrestasi string,
) (model.Prestasi, error) {
	var prestasi model.Prestasi
	var namaPrestasi string
	var namaMahasiswa *string

	query := `
        SELECT
            p.id_prestasi,
            p.nama_prestasi,
            p.juara,
            p.nim,
            s.name
        FROM prestasi p
        LEFT JOIN students s ON s.nim = p.nim
        WHERE p.id_prestasi = $1
    `

	err := r.pool.QueryRow(ctx, query, idPrestasi).Scan(
		&prestasi.IDPrestasi,
		&namaPrestasi,
		&prestasi.Juara,
		&prestasi.NIM,
		&namaMahasiswa,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Prestasi{}, ErrPrestasiNotFound
		}

		return model.Prestasi{}, err
	}

	prestasi.NamaPrestasi = strings.TrimSpace(namaPrestasi)
	prestasi.NamaMhs = namaMahasiswa

	return prestasi, nil
}
