package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bank-repayment-record/backend/internal/config"
	"bank-repayment-record/backend/internal/repayment"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	db  *sql.DB
	loc *time.Location
}

func OpenMySQL(cfg config.MySQLConfig, loc *time.Location) (*MySQLStore, error) {
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, fmt.Errorf("build mysql dsn: %w", err)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return &MySQLStore{db: db, loc: loc}, nil
}

func OpenMySQLDSN(dsn string, loc *time.Location) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return &MySQLStore{db: db, loc: loc}, nil
}

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) CreateRepayment(ctx context.Context, record repayment.Record) (repayment.Record, error) {
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
		repaymentAt,
		now,
		now,
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

func (s *MySQLStore) DeleteRepayment(ctx context.Context, id int64) (bool, error) {
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

func (s *MySQLStore) ListRepayments(ctx context.Context, filters repayment.Filters) ([]repayment.Record, error) {
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
			record                            repayment.Record
			repaymentAt, createdAt, updatedAt time.Time
		)
		if err := rows.Scan(
			&record.ID,
			&record.CardName,
			&record.Currency,
			&record.AmountCents,
			&repaymentAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan repayment row: %w", err)
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

func (s *MySQLStore) ClearRepayments(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM repayments`); err != nil {
		return fmt.Errorf("clear repayments: %w", err)
	}
	return nil
}
