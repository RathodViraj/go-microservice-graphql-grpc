package order

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error)
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

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) (err error) {
	query := `INSERT INTO orders (id, created_at, account_id, total_price) VALUES ($1, $2, $3, $4)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.ExecContext(
		ctx,
		query,
		o.ID,
		o.CreatedAt,
		o.AccountID,
		o.TotalPrice,
	)
	if err != nil {
		return
	}

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("orders_products", "order_id", "product_id", "quantity"))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return
		}
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	return
}

func (r *postgresRepository) GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error) {
	query := `
		SELECT o.id,
		       o.created_at,
		       o.account_id,
		       o.total_price::numeric::float8,
		       op.product_id,
		       op.quantity
		FROM orders o
		JOIN orders_products op ON (o.id = op.order_id)
		WHERE o.account_id = $1
		ORDER BY o.id
	`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ordersMap := map[string]*Order{}
	orderIDs := []string{}

	for rows.Next() {
		var id, accID, productID string
		var createdAt pq.NullTime
		var totalPrice float64
		var quantity int64

		if err := rows.Scan(&id, &createdAt, &accID, &totalPrice, &productID, &quantity); err != nil {
			return nil, err
		}

		ord, ok := ordersMap[id]
		if !ok {
			ord = &Order{
				ID:         id,
				CreatedAt:  createdAt.Time,
				AccountID:  accID,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			ordersMap[id] = ord
			orderIDs = append(orderIDs, id)
		}

		ord.Products = append(ord.Products, OrderedProduct{
			ID:       productID,
			Quantity: uint32(quantity),
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	orders := []Order{}
	for _, id := range orderIDs {
		if ord, ok := ordersMap[id]; ok {
			orders = append(orders, *ord)
		}
	}

	return orders, nil
}
