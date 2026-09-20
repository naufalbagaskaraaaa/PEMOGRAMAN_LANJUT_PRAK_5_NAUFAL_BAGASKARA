package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository interface {
	LoadPermissions(ctx context.Context) (map[string][]string, error)
}

type rolePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) RoleRepository {
	return &rolePostgresRepository{pool: pool}
}

func (r *rolePostgresRepository) LoadPermissions(ctx context.Context) (map[string][]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.name, COALESCE(rp.permission_name, '')
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_name = r.name
		ORDER BY r.name, rp.permission_name
	`)
	if err != nil {
		return nil, fmt.Errorf("memuat permission role: %w", err)
	}
	defer rows.Close()

	permissionsByRole := make(map[string][]string)
	for rows.Next() {
		var roleName, permissionName string
		if err := rows.Scan(&roleName, &permissionName); err != nil {
			return nil, fmt.Errorf("membaca permission role: %w", err)
		}

		if _, exists := permissionsByRole[roleName]; !exists {
			permissionsByRole[roleName] = []string{}
		}
		if permissionName != "" {
			permissionsByRole[roleName] = append(permissionsByRole[roleName], permissionName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca daftar permission role: %w", err)
	}

	return permissionsByRole, nil
}
