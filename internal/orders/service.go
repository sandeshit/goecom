package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
)

var (
	ErrProductNotFound = errors.New("Product not found")
	ErrOutOfStock      = errors.New("Product out of stock")
)

type svc struct {
	repo *repo.Queries
	db   *pgx.Conn
}

func NewService(repo *repo.Queries, db *pgx.Conn) Service {
	return &svc{
		repo: repo,
		db:   db,
	}
}

func (s *svc) PlaceOrder(ctx context.Context, tempOrder createOrderParams) (repo.Order, error) {
	if tempOrder.CustomerID == 0 {
		return repo.Order{}, fmt.Errorf("customer ID is required")
	}

	if len(tempOrder.Items) == 0 {
		return repo.Order{}, fmt.Errorf("Atleast one item is required")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}

	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	order, err := qtx.CreateOrder(ctx, tempOrder.CustomerID)
	if err != nil {
		return repo.Order{}, ErrProductNotFound
	}

	for _, item := range tempOrder.Items {
		product, err := qtx.FindProductByID(ctx, item.ProductID)
		if err != nil {
			return repo.Order{}, err
		}
		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrOutOfStock
		}

		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:        order.ID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			PriceInCenters: product.PriceInCenters,
		})
		if err != nil {
			return repo.Order{}, err
		}
		// Challenge: update the product stock quantity.
	}

	tx.Commit(ctx)
	return order, nil
}
