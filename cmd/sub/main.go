package main

import (
	"encoding/json"
	"fmt"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/repository"
	"time"

	"github.com/MikhailKatarzhin/Level0Golang/internal/database/postgre"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/stan"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	addr    = "localhost:4222"
	CID     = "clientID"
	clstrID = "wbl0ns"
	consmr  = "testConsumer"
	user    = "wbl0user"
	pass    = "wbl0pass"
	subject = "testing"
)

func main() {

	pgConnPool, err := postgre.DefaultCredConfig()

	if err != nil {
		panic(err)
	}

	defer pgConnPool.Close()

	logger.L().Info("Successfully connected to postgres")

	orderServ := orderService.NewOrderService(repository.NewOrderRepository(pgConnPool))
	logger.L().Info("Successfully created repository and service for postgres")

	//TODO pull orders from BD to cache

	client := stan.New(broker.NATSConfig{
		Addr:     addr,
		User:     user,
		Password: pass,
	})

	if err := client.Connect(clstrID, fmt.Sprint(CID, "-", gonanoid.Must(5))); err != nil {
		panic(err.Error())
	}

	defer func(client *stan.Client) {
		if err := client.Disconnect(); err != nil {
			panic(err.Error())
		}
	}(client)

	subs, err := client.QueueSubscribeWithAck(subject, consmr)

	if err != nil {
		panic(err.Error())
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-done:

				if err := subs.Unsubscribe.Unsubscribe(); err != nil {
					panic(err.Error())
				}

			case data := <-subs.Ch:

				if err := data.AckCallback(); err != nil {
					panic(err.Error())
				}

				//TODO check received JSON with JSON schema

				receivedOrder, err := UnmarshalOrder(data.Body)
				if err != nil {
					println(err)
				}

				logger.L().Info(fmt.Sprintf("Received STAN order: [uid]%s", receivedOrder.OrderUID))

				if err := orderServ.InsertOrderToDB(receivedOrder); err != nil {
					logger.L().Error(fmt.Sprintf(
						"Failed to insert order[uid:%s] to DB: %s",
						receivedOrder.OrderUID,
						err.Error(),
					))
				} else {
					logger.L().Info(fmt.Sprintf("Successful insert order[uid:%s] to BD", receivedOrder.OrderUID))
				}
				//TODO insert received and uploaded order into cache
			}
		}
	}()

	time.Sleep(30 * time.Minute)
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
