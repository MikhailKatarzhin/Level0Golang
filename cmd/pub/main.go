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

	if err := client.Publish(subject, []byte("Hello, NATS Streaming!")); err != nil {
		panic(err.Error())
	}

	time.Sleep(300 * time.Second)
}
