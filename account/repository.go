package account

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrAccountNotFound = errors.New("account not found")

type Repository interface {
	Close()
	PutAccount(ctx context.Context, a Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip, take uint64) ([]Account, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepository) PutAccount(ctx context.Context, a Account) error {
	query := `INSERT INTO accounts(id, name) VALUES($1,$2)`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.Name)
	return err
}

func (r *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	query := `SELECT id, name FROM accounts WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	a := &Account{}
	if err := row.Scan(&a.ID, &a.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}

	return a, nil
}

func (r *postgresRepository) ListAccounts(ctx context.Context, skip, take uint64) ([]Account, error) {
	query := `SELECT id, name FROM accounts ORDER BY id DESC OFFSET $1 LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, skip, take)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err = rows.Scan(&a.ID, &a.Name); err == nil {
			accounts = append(accounts, *a)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}
