package main

import (
	"encoding/json"
	"fmt"
	"github.com/MikhailKatarzhin/Level0Golang/internal/database/postgre/repository"
	"time"

	"github.com/MikhailKatarzhin/Level0Golang/internal/database/postgre"
	"github.com/MikhailKatarzhin/Level0Golang/internal/order"
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

	orderRepo := repository.NewOrderRepository(pgConnPool)
	logger.L().Info("Successfully created repository for postgres")

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

				receivedOrder, err := UnmarshalTheMessage(data.Body)
				if err != nil {
					println(err)
				}

				logger.L().Info(fmt.Sprintf("Received STAN order: [uid]%s", receivedOrder.OrderUID))

				// Inserting order into bd
				if err := orderRepo.InsertOrderToDB(receivedOrder); err != nil {
					logger.L().Error(fmt.Sprintf(
						"Failed to insert order[uid:%s] to DB: %s",
						receivedOrder.OrderUID,
						err.Error(),
					))
				} else {
					logger.L().Info(fmt.Sprintf("Successful insert order[uid:%s] to BD", receivedOrder.OrderUID))
				}
			}
		}
	}()

	time.Sleep(30 * time.Minute)
}

func UnmarshalTheMessage(dataByte []byte) (order.Order, error) {
	var newOrder order.Order
	err := json.Unmarshal(dataByte, &newOrder)

	if err != nil {
		logger.L().Error(fmt.Sprintf("error while unmarshalling message to model : %s", err.Error()))
		return newOrder, err
	}

	return newOrder, nil
}
