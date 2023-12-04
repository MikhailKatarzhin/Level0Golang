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

func (repo *OrderRepository) GetOrderByUID(orderUID string) (order.Order, error) {
	var ordr order.Order

	orderQuery := `
		SELECT order_uid, track_number, entry, locale, internal_signature, 
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders 
		WHERE order_uid = $1
	`
	row := repo.pgConnPool.QueryRow(context.Background(), orderQuery, orderUID)
	err := row.Scan(
		&ordr.OrderUID,
		&ordr.TrackNumber,
		&ordr.Entry,
		&ordr.Locale,
		&ordr.InternalSignature,
		&ordr.CustomerId,
		&ordr.DeliveryService,
		&ordr.ShardKey,
		&ordr.SmId,
		&ordr.DateCreated,
		&ordr.OofShard,
	)
	if err != nil {
		return ordr, err
	}

	deliveryQuery := `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1
	`
	row = repo.pgConnPool.QueryRow(context.Background(), deliveryQuery, orderUID)
	err = row.Scan(

		&ordr.Delivery.Name,
		&ordr.Delivery.Phone,
		&ordr.Delivery.Zip,
		&ordr.Delivery.City,
		&ordr.Delivery.Address,
		&ordr.Delivery.Region,
		&ordr.Delivery.Email,
	)
	if err != nil {
		return ordr, err
	}

	paymentQuery := `
		SELECT transaction, request_id, currency, provider, amount, 
		       payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE transaction = $1
	`
	row = repo.pgConnPool.QueryRow(context.Background(), paymentQuery, orderUID)
	err = row.Scan(
		&ordr.Payment.Transaction,
		&ordr.Payment.RequestId,
		&ordr.Payment.Currency,
		&ordr.Payment.Provider,
		&ordr.Payment.Amount,
		&ordr.Payment.PaymentDt,
		&ordr.Payment.Bank,
		&ordr.Payment.DeliveryCost,
		&ordr.Payment.GoodsTotal,
		&ordr.Payment.CustomFee,
	)
	if err != nil {
		return ordr, err
	}

	itemsQuery := `
		SELECT chrt_id, track_number, price, rid, name, 
		       sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE track_number = $1
	`
	rows, err := repo.pgConnPool.Query(context.Background(), itemsQuery, ordr.TrackNumber)
	if err != nil {
		return ordr, err
	}
	defer rows.Close()

	for rows.Next() {
		var item order.Item
		err := rows.Scan(
			&item.ChrtId,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmId,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return ordr, err
		}
		ordr.Items = append(ordr.Items, item)
	}

	return ordr, nil
}

//TODO download all orders from BD by uid
