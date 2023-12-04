package main

import (
	"fmt"
	"time"

	"github.com/MikhailKatarzhin/Level0Golang/internal/database/postgre"
	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService"
	//"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/model"
	//"github.com/MikhailKatarzhin/Level0Golang/internal/orderService/repository"
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

	//orderServ := orderService.NewOrderService(repository.NewOrderRepository(pgConnPool))
	//logger.L().Info("Successfully created repository and service for postgres")

	//TODO pull orders from BD to cache

	lruCache := cache.NewLRUCache[string, []byte](3600)

	jobQueue := make(chan []byte, 100)

	orderService.StartWorkerPool(nWorker, jobQueue, pgConnPool, lruCache)

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
