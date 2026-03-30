package repository

import (
	"context"
	"database/sql"
)

type Policy struct {
	ID          uint64
	Version     string
	ContentHash string
}

type PolicyRepository interface {
	List(ctx context.Context, limit int) ([]Policy, error)
	Create(ctx context.Context, p Policy) (*Policy, error)
}

type mysqlPolicyRepository struct {
	db *sql.DB
}

func NewPolicyRepository(db *sql.DB) PolicyRepository {
	return &mysqlPolicyRepository{db: db}
}

func (r *mysqlPolicyRepository) Create(ctx context.Context, p Policy) (*Policy, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO policies (version, content_hash) VALUES (?, ?)`,
		p.Version, p.ContentHash,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	p.ID = uint64(id)
	return &p, nil
}

func (r *mysqlPolicyRepository) List(ctx context.Context, limit int) ([]Policy, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, version, content_hash
		FROM policies
		ORDER BY id
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	policies := make([]Policy, 0)
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Version, &p.ContentHash); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}
