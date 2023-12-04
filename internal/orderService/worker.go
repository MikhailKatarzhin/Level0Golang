package orderService

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/repository"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"
	"github.com/jackc/pgx/v4/pgxpool"
)

func StartWorkerPool(numWorkers int, jobQueue chan []byte, pgConnPool *pgxpool.Pool) {
	orderServ := NewOrderService(repository.NewOrderRepository(pgConnPool))

	for i := 0; i < numWorkers; i++ {
		go work(i, jobQueue, orderServ)
	}
}

func work(workerID int, jobQueue chan []byte, orderServ *OrderService) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultWorkerCtxTimeout)
	defer cancel()

	for data := range jobQueue {

		//TODO check received JSON with JSON schema

		order, err := UnmarshalOrder(data)
		if err != nil {
			println(err)
		}

		logger.L().Info(fmt.Sprintf("[%d]Received STAN order: [uid]%s",
			workerID,
			order.OrderUID,
		))

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
		}
		//TODO insert received and uploaded order into cache
	}
}

func UnmarshalOrder(dataByte []byte) (model.Order, error) {
	var newOrder model.Order
	err := json.Unmarshal(dataByte, &newOrder)

	if err != nil {
		logger.L().Error(fmt.Sprintf("error while unmarshalling message to order : %s", err.Error()))
		return newOrder, err
	}

	return newOrder, nil
}

func MarshalOrder(ordr model.Order) ([]byte, error) {
	dataByte, err := json.Marshal(ordr)

	if err != nil {
		logger.L().Error(fmt.Sprintf("error while marshalling order to databyte : %s", err.Error()))
		return dataByte, err
	}

	return dataByte, nil
}
