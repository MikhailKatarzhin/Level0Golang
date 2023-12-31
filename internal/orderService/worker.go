package orderService

import (
	"context"
	"fmt"

	cchr "github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/cache"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/posgre"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/cache"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"

	"github.com/jackc/pgx/v4/pgxpool"
)

func StartWorkerPool(numWorkers int, jobQueue chan []byte, pgConnPool *pgxpool.Pool, cache *cache.LRUCache[string, []byte]) {
	orderServ := NewOrderService(posgre.NewOrderRepository(pgConnPool), cchr.NewOrderRepository(cache))
	StartWorkerPoolWithOrderService(numWorkers, jobQueue, orderServ)
}

func StartWorkerPoolWithOrderService(numWorkers int, jobQueue chan []byte, orderServ *OrderService) {
	for i := 0; i < numWorkers; i++ {
		go work(i, jobQueue, orderServ)
	}
}

func work(workerID int, jobQueue chan []byte, orderServ *OrderService) {
	for data := range jobQueue {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultWorkerCtxTimeout)
		defer cancel()

		//TODO check received JSON with JSON schema

		order, err := UnmarshalOrder(data)
		if err != nil {
			println(err)
		}

		logger.L().Info(fmt.Sprintf("[%d]Received STAN order: [uid:%s]",
			workerID,
			order.OrderUID,
		))

		if _, exist := orderServ.CacheRepo.GetOrderByUID(order.OrderUID); exist {
			if order, err := orderServ.GetOrderByOrderUIDFromBD(order.OrderUID); err != nil {
				setToCache(order.OrderUID, data, workerID, orderServ)
			} else {
				logger.L().Warn(fmt.Sprintf("[%d]Received order[uid:%s] alrady exist",
					workerID,
					order.OrderUID,
				))
				continue
			}
		}

		if err := orderServ.InsertOrderToDB(ctx, order); err != nil {
			logger.L().Error(fmt.Sprintf(
				"[%d]Failed to insert order[uid:%s] to DB: %s",
				workerID,
				order.OrderUID,
				err.Error(),
			))
		} else {
			logger.L().Info(fmt.Sprintf("[%d]Successful insert order[uid:%s] to BD",
				workerID,
				order.OrderUID,
			))
			setToCache(order.OrderUID, data, workerID, orderServ)
		}
	}
}

func setToCache(orderUID string, data []byte, workerID int, orderServ *OrderService) {
	orderServ.InsertOrderToCache(orderUID, data)
	logger.L().Info(fmt.Sprintf("[%d]Successful insert order[uid:%s] to cache",
		workerID,
		orderUID,
	))
}
