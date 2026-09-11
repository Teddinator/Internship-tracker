package applications

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) GetByID(
	ctx context.Context,
	id int64,
) (Application, error) {
	var a Application

	err := s.db.QueryRowContext(
		ctx,
		`
			SELECT 
				a.id,
				a.company_id,
				c.name,
				a.role,
				a.status,
				a.applied_at,
				a.created_at,
				a.updated_at
			FROM applications AS a
			JOIN companies AS c
				ON c.id = a.company_id
			WHERE a.id = $1
		`, id,
	).Scan(
		&a.ID,
		&a.CompanyID,
		&a.Company,
		&a.Role,
		&a.Status,
		&a.AppliedAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		return Application{}, err
	}

	return a, nil
}

func (s *PostgresStore) Delete(
	ctx context.Context,
	id int64,
) (bool, error) {
	result, err := s.db.ExecContext(
		ctx,
		`
			DELETE FROM applications
			WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return false, err
	}

	rowsaffected, err := result.RowsAffected()

	if err != nil {
		return false, err
	}

	return rowsaffected > 0, nil
}

func (s *PostgresStore) GetAll(
	ctx context.Context,
	filter applicationFilter,
) ([]Application, error) {
	query := `
		SELECT
			a.id,
			a.company_id,
			c.name,
			a.role,
			a.status,
			a.applied_at,
			a.created_at,
			a.updated_at 
		FROM applications AS a
		JOIN companies AS c
			ON c.id = a.company_id
		WHERE 1 = 1
	`

	args := make([]any, 0)

	if filter.Status != "" {
		args = append(args, filter.Status)
		query += fmt.Sprintf(" AND a.status = $%d", len(args))

	}

	if filter.CompanyID != "" {
		companyID, err := uuid.Parse(filter.CompanyID)
		if err != nil {
			return nil, err
		}

		args = append(args, companyID)
		query += fmt.Sprintf(" AND a.company_id = $%d", len(args))
	}

	if filter.Location != "" {
		args = append(args, "%"+filter.Location+"%")
		query += fmt.Sprintf(" AND c.location ILIKE $%d", len(args))
	}

	if filter.Industry != "" {
		args = append(args, "%"+filter.Industry+"%")
		query += fmt.Sprintf(" AND c.industry ILIKE $%d", len(args))
	}

	query += " ORDER BY a.id"

	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applications := make([]Application, 0)

	for rows.Next() {
		var app Application

		if err := rows.Scan(
			&app.ID,
			&app.CompanyID,
			&app.Company,
			&app.Role,
			&app.Status,
			&app.AppliedAt,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
			return nil, err
		}

		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}
