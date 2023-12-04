package posgre

import (
	"context"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository struct {
	PgConnPool *pgxpool.Pool
}

func NewOrderRepository(pgConnPool *pgxpool.Pool) *Repository {
	return &Repository{PgConnPool: pgConnPool}
}

func (repo *Repository) InsertOrder(ctx context.Context, newOrder model.Order) error {
	query := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := repo.PgConnPool.Exec(ctx, query,
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

func (repo *Repository) InsertDelivery(ctx context.Context, newOrder model.Order) error {
	query := `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := repo.PgConnPool.Exec(ctx, query,
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

func (repo *Repository) InsertPayment(ctx context.Context, newOrder model.Order) error {
	query := `
		INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := repo.PgConnPool.Exec(ctx, query,
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

func (repo *Repository) InsertItem(ctx context.Context, newItem model.Item) error {
	query := `
		INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := repo.PgConnPool.Exec(ctx, query,
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

func (repo *Repository) GetOrderByUID(orderUID string) (model.Order, error) {
	var ordr model.Order

	orderQuery := `
		SELECT order_uid, track_number, entry, locale, internal_signature, 
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders 
		WHERE order_uid = $1
	`
	row := repo.PgConnPool.QueryRow(context.Background(), orderQuery, orderUID)
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
	row = repo.PgConnPool.QueryRow(context.Background(), deliveryQuery, orderUID)
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
	row = repo.PgConnPool.QueryRow(context.Background(), paymentQuery, orderUID)
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
	rows, err := repo.PgConnPool.Query(context.Background(), itemsQuery, ordr.TrackNumber)
	if err != nil {
		return ordr, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
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

func (repo *Repository) GetAllOrders() ([]model.Order, error) {
	var orders []model.Order

	uids, err := repo.GetAllOrderUIDs()
	if err != nil {
		return nil, err
	}

	for _, uid := range uids {
		orderData, err := repo.GetOrderByUID(uid)
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderData)
	}

	return orders, nil
}

func (repo *Repository) GetAllOrderUIDs() ([]string, error) {
	var orderUIDs []string

	orderUIDQuery := `SELECT order_uid FROM orders`

	rows, err := repo.PgConnPool.Query(context.Background(), orderUIDQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderUID string
		err := rows.Scan(&orderUID)
		if err != nil {
			return nil, err
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	return orderUIDs, nil
}
