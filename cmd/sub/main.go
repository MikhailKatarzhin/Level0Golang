package main

import (
	"fmt"
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
				println(string(data.Body))

				if err := data.AckCallback(); err != nil {
					panic(err.Error())
				}
			}
		}
	}()

	time.Sleep(30 * time.Minute)
}
