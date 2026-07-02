package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bank-repayment-record/backend/internal/repayment"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db  *sql.DB
	loc *time.Location
}

func OpenSQLite(dbPath string, loc *time.Location) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &SQLiteStore{db: db, loc: loc}
	if err := store.initSchema(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) CreateRepayment(ctx context.Context, record repayment.Record) (repayment.Record, error) {
	query := `
		INSERT INTO repayments (card_name, currency, amount, repayment_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now().In(s.loc).Truncate(time.Second)
	repaymentAt := record.RepaymentAt.In(s.loc).Truncate(time.Minute)
	res, err := s.db.ExecContext(
		ctx,
		query,
		record.CardName,
		record.Currency,
		record.AmountCents,
		repaymentAt.Format(time.RFC3339),
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return repayment.Record{}, fmt.Errorf("insert repayment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return repayment.Record{}, fmt.Errorf("last insert id: %w", err)
	}

	record.ID = id
	record.RepaymentAt = repaymentAt
	record.CreatedAt = now
	record.UpdatedAt = now

	return record, nil
}

func (s *SQLiteStore) DeleteRepayment(ctx context.Context, id int64) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM repayments WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("delete repayment: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}

	return rows > 0, nil
}

func (s *SQLiteStore) ListRepayments(ctx context.Context, filters repayment.Filters) ([]repayment.Record, error) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)

	if filters.CardName != "" {
		conditions = append(conditions, "card_name = ?")
		args = append(args, filters.CardName)
	}
	if filters.Currency != "" {
		conditions = append(conditions, "currency = ?")
		args = append(args, filters.Currency)
	}

	query := `
		SELECT id, card_name, currency, amount, repayment_at, created_at, updated_at
		FROM repayments
	`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY repayment_at DESC, id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query repayments: %w", err)
	}
	defer rows.Close()

	records := make([]repayment.Record, 0)
	for rows.Next() {
		var (
			record                       repayment.Record
			repaymentAtStr, createdAtStr string
			updatedAtStr                 string
		)
		if err := rows.Scan(
			&record.ID,
			&record.CardName,
			&record.Currency,
			&record.AmountCents,
			&repaymentAtStr,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan repayment row: %w", err)
		}

		repaymentAt, err := time.Parse(time.RFC3339, repaymentAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse repayment_at for id %d: %w", record.ID, err)
		}
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at for id %d: %w", record.ID, err)
		}
		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at for id %d: %w", record.ID, err)
		}

		record.RepaymentAt = repaymentAt.In(s.loc)
		record.CreatedAt = createdAt.In(s.loc)
		record.UpdatedAt = updatedAt.In(s.loc)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repayment rows: %w", err)
	}

	return records, nil
}

func (s *SQLiteStore) initSchema(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS repayments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	card_name TEXT NOT NULL,
	currency TEXT NOT NULL,
	amount INTEGER NOT NULL,
	repayment_at TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repayments_repayment_at ON repayments (repayment_at);
CREATE INDEX IF NOT EXISTS idx_repayments_currency ON repayments (currency);
CREATE INDEX IF NOT EXISTS idx_repayments_card_name ON repayments (card_name);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}
	return nil
}
