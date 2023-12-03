package repository

import (
	"context"
	"github.com/MikhailKatarzhin/Level0Golang/internal/order"
	"github.com/jackc/pgx/v4/pgxpool"
)

type OrderRepository struct {
	pgConnPool *pgxpool.Pool
}

func NewOrderRepository(pgConnPool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pgConnPool: pgConnPool}
}

func (repo *OrderRepository) InsertOrderToDB(newOrder order.Order) error {
	tx, err := repo.pgConnPool.Begin(context.Background())
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
			panic(p)
		}
	}()

	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		}
	}()

	if err := repo.InsertOrder(newOrder); err != nil {
		return err
	}

	if err := repo.InsertDelivery(newOrder); err != nil {
		return err
	}

	if err := repo.InsertPayment(newOrder); err != nil {
		return err
	}

	for _, item := range newOrder.Items {
		if err := repo.InsertItem(item); err != nil {
			return err
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		return err
	}
	return nil
}

func (repo *OrderRepository) InsertOrder(newOrder order.Order) error {
	query := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := repo.pgConnPool.Exec(context.Background(), query,
		newOrder.OrderUID,
		newOrder.TrackNumber,
		newOrder.Entry,
		newOrder.Locale,
		newOrder.InternalSignature,
		newOrder.CustomerId,
		newOrder.DeliveryService,
		newOrder.ShardKey,
		newOrder.SmId,
		newOrder.DateCreated,
		newOrder.OofShard,
	)

	return err
}

func (repo *OrderRepository) InsertDelivery(newOrder order.Order) error {
	query := `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := repo.pgConnPool.Exec(context.Background(), query,
		newOrder.OrderUID,
		newOrder.Delivery.Name,
		newOrder.Delivery.Phone,
		newOrder.Delivery.Zip,
		newOrder.Delivery.City,
		newOrder.Delivery.Address,
		newOrder.Delivery.Region,
		newOrder.Delivery.Email,
	)

	return err
}

func (repo *OrderRepository) InsertPayment(newOrder order.Order) error {
	query := `
		INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := repo.pgConnPool.Exec(context.Background(), query,
		newOrder.Payment.Transaction,
		newOrder.Payment.RequestId,
		newOrder.Payment.Currency,
		newOrder.Payment.Provider,
		newOrder.Payment.Amount,
		newOrder.Payment.PaymentDt,
		newOrder.Payment.Bank,
		newOrder.Payment.DeliveryCost,
		newOrder.Payment.GoodsTotal,
		newOrder.Payment.CustomFee,
	)

	return err
}

func (repo *OrderRepository) InsertItem(newItem order.Item) error {
	query := `
		INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := repo.pgConnPool.Exec(context.Background(), query,
		newItem.ChrtId,
		newItem.TrackNumber,
		newItem.Price,
		newItem.Rid,
		newItem.Name,
		newItem.Sale,
		newItem.Size,
		newItem.TotalPrice,
		newItem.NmId,
		newItem.Brand,
		newItem.Status,
	)

	return err
}
