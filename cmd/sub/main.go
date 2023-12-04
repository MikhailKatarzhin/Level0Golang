package main

import (
	"fmt"
	cchr "github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/cache"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model/posgre"
	"time"

	"github.com/MikhailKatarzhin/Level0Golang/internal/database/postgre"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/stan"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/cache"
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
	nWorker = 10
)

func main() {

	pgConnPool, err := postgre.DefaultCredConfig()

	if err != nil {
		panic(err)
	}

	defer pgConnPool.Close()

	logger.L().Info("Successfully connected to postgres")

	lruCache := cache.NewLRUCache[string, []byte](3600)
	orderServ := orderService.NewOrderService(posgre.NewOrderRepository(pgConnPool), cchr.NewOrderRepository(lruCache))
	logger.L().Info("Successfully created repositories and service for postgres")

	err = orderServ.LoadAllOrdersToCacheFromBD()
	if err != nil {
		logger.L().Info(fmt.Sprintf("During loading orders BD to cache was errors:%s", err.Error()))
	}

	jobQueue := make(chan []byte, 100)

	orderService.StartWorkerPoolWithOrderService(nWorker, jobQueue, orderServ)

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

				jobQueue <- data.Body
			}
		}
	}()

	time.Sleep(30 * time.Minute)
}
