package main

import (
	"fmt"
	"time"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/stan"

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
			}
		}
	}()

	time.Sleep(30 * time.Minute)
}
