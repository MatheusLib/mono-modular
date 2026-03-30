package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Consent struct {
	ID        uint64
	UserID    uint64
	PolicyID  uint64
	Purpose   string
	Status    string
	CreatedAt time.Time
	RevokedAt *time.Time
}

type ConsentRepository interface {
	List(ctx context.Context, limit int) ([]Consent, error)
	Create(ctx context.Context, c Consent) (*Consent, error)
	Revoke(ctx context.Context, documentID uint64) error
}

type mysqlConsentRepository struct {
	db *sql.DB
}

func NewConsentRepository(db *sql.DB) ConsentRepository {
	return &mysqlConsentRepository{db: db}
}

func (r *mysqlConsentRepository) Create(ctx context.Context, c Consent) (*Consent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO consents (user_id, policy_id, purpose, status) VALUES (?, ?, ?, 'active')`,
		c.UserID, c.PolicyID, c.Purpose,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	payload := fmt.Sprintf(`{"consent_id":%d}`, id)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO audit_events (event_type, entity_type, entity_id, payload_json) VALUES ('ConsentCreated', 'consent', ?, ?)`,
		id, payload,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	c.ID = uint64(id)
	c.Status = "active"
	return &c, nil
}

func (r *mysqlConsentRepository) Revoke(ctx context.Context, documentID uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`UPDATE consents SET status='revoked', revoked_at=NOW() WHERE id=? AND status='active'`,
		documentID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	payload := fmt.Sprintf(`{"consent_id":%d}`, documentID)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO audit_events (event_type, entity_type, entity_id, payload_json) VALUES ('ConsentRevoked', 'consent', ?, ?)`,
		documentID, payload,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *mysqlConsentRepository) List(ctx context.Context, limit int) ([]Consent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, policy_id, purpose, status
		FROM consents
		ORDER BY id
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	consents := make([]Consent, 0)
	for rows.Next() {
		var c Consent
		if err := rows.Scan(&c.ID, &c.UserID, &c.PolicyID, &c.Purpose, &c.Status); err != nil {
			return nil, err
		}
		consents = append(consents, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return consents, nil
}
