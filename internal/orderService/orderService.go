package orderService

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/cache"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/posgre"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

type OrderService struct {
	PostgresRepo *posgre.Repository
	CacheRepo    *cache.Repository
}

func NewOrderService(orderRepo *posgre.Repository, cache *cache.Repository) *OrderService {
	return &OrderService{PostgresRepo: orderRepo, CacheRepo: cache}
}

func (osrv *OrderService) InsertOrderToDB(ctx context.Context, newOrder model.Order) error {

	tx, err := osrv.PostgresRepo.PgConnPool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := osrv.PostgresRepo.InsertOrder(ctx, newOrder); err != nil {
		return fmt.Errorf("failure insert order by [uid:%s] into bd during insert order: %s",
			newOrder.OrderUID, err)
	}

	if err := osrv.PostgresRepo.InsertDelivery(ctx, newOrder); err != nil {
		return fmt.Errorf("failure insert order by [uid:%s] into bd during insert delivery: %s",
			newOrder.OrderUID, err,
		)
	}

	if err := osrv.PostgresRepo.InsertPayment(ctx, newOrder); err != nil {
		return fmt.Errorf("failure insert order by [uid:%s] into bd during insert payment: %s",
			newOrder.OrderUID, err,
		)
	}

	for _, item := range newOrder.Items {
		if err := osrv.PostgresRepo.InsertItem(ctx, item); err != nil {
			return fmt.Errorf("failure insert order by [uid:%s] into bd during insert item: %s",
				newOrder.OrderUID, err,
			)
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failure insert order by [uid:%s] into bd during commit transaction: %s",
			newOrder.OrderUID, err,
		)
	}
	return nil
}

func (osrv *OrderService) GetOrderByOrderUIDFromBD(orderUID string) (model.Order, error) {
	var selectedOrder model.Order

	selectedOrder, err := osrv.PostgresRepo.GetOrderByUID(orderUID)
	if err != nil {
		logger.L().Error(fmt.Sprintf("Failure selecting order by [uid:%s] from bd: %s", orderUID, err.Error()))
		return selectedOrder, err
	}
	logger.L().Info(fmt.Sprintf("Suscessful selecting order by [uid:%s] from bd", orderUID))
	return selectedOrder, nil
}

func (osrv *OrderService) InsertOrderToCache(orderUID string, data []byte) {
	osrv.CacheRepo.InsertOrder(orderUID, data)
}

func (osrv *OrderService) GetOrderByOrderUIDFromCache(orderUID string) ([]byte, bool) {
	return osrv.CacheRepo.GetOrderByUID(orderUID)
}

func (osrv *OrderService) LoadAllOrdersFromBD() ([]model.Order, error) {
	return osrv.PostgresRepo.GetAllOrders()
}

func (osrv *OrderService) LoadAllOrdersToCacheFromBD() error {
	orders, err := osrv.LoadAllOrdersFromBD()
	if err != nil {
		return fmt.Errorf("error during loads all orders from BD: %s", err.Error())
	}

	var sumErr error

	for i, order := range orders {
		orderByte, err := MarshalOrder(order)
		if err != nil {
			sumErr = fmt.Errorf("error during marshal order[%d]: %s; ", i, sumErr.Error())
			continue
		}

		osrv.InsertOrderToCache(order.OrderUID, orderByte)
	}

	if sumErr != nil {
		return sumErr
	}

	return nil
}

func UnmarshalOrder(dataByte []byte) (model.Order, error) {
	var newOrder model.Order
	err := json.Unmarshal(dataByte, &newOrder)

	if err != nil {
		return newOrder, fmt.Errorf("error while unmarshalling message to order : %s", err.Error())
	}

	return newOrder, nil
}

func MarshalOrder(ordr model.Order) ([]byte, error) {
	dataByte, err := json.Marshal(ordr)

	if err != nil {
		return dataByte, fmt.Errorf("error while marshalling order to databyte : %s", err.Error())
	}

	return dataByte, nil
}
