package orderService

import (
	"context"
	"fmt"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"

	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/repository"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
)

type OrderService struct {
	OrderRepo *repository.OrderRepository
}

func NewOrderService(orderRepo *repository.OrderRepository) *OrderService {
	return &OrderService{OrderRepo: orderRepo}
}

func (osrv *OrderService) InsertOrderToDB(ctx context.Context, newOrder model.Order) error {

	tx, err := osrv.OrderRepo.PgConnPool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := osrv.OrderRepo.InsertOrder(ctx, newOrder); err != nil {
		logger.L().Error(fmt.Sprintf("Failure insert order by [uid:%s] into bd during insert order: %s",
			newOrder.OrderUID, err.Error(),
		))
		return err
	}

	if err := osrv.OrderRepo.InsertDelivery(ctx, newOrder); err != nil {
		logger.L().Error(fmt.Sprintf("Failure insert order by [uid:%s] into bd during insert delivery: %s",
			newOrder.OrderUID, err.Error(),
		))
		return err
	}

	if err := osrv.OrderRepo.InsertPayment(ctx, newOrder); err != nil {
		logger.L().Error(fmt.Sprintf("Failure insert order by [uid:%s] into bd during insert payment: %s",
			newOrder.OrderUID, err.Error(),
		))
		return err
	}

	for _, item := range newOrder.Items {
		if err := osrv.OrderRepo.InsertItem(ctx, item); err != nil {
			logger.L().Error(fmt.Sprintf("Failure insert order by [uid:%s] into bd during insert item: %s",
				newOrder.OrderUID, err.Error(),
			))
			return err
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		logger.L().Error(fmt.Sprintf("Failure insert order by [uid:%s] into bd during commit transaction: %s",
			newOrder.OrderUID, err.Error(),
		))
		return err
	}
	return nil
}

func (osrv *OrderService) GetOrderByOrderUID(orderUID string) (model.Order, error) {
	var selectedOrder model.Order

	selectedOrder, err := osrv.OrderRepo.GetOrderByUID(orderUID)
	if err != nil {
		logger.L().Error(fmt.Sprintf("Failure selecting order by [uid:%s] from bd: %s", orderUID, err.Error()))
		return selectedOrder, err
	}
	logger.L().Info(fmt.Sprintf("Suscessful selecting order by [uid:%s] from bd", orderUID))
	return selectedOrder, err
}
