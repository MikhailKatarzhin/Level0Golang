package main

import (
	"fmt"
	"io"
	"os"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/stan"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	addr    = "localhost:4222"
	CID     = "clientID"
	clstrID = "wbl0ns"
	jsonF   = "cmd/pub/test.json"
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

	// Чтение файла JSON
	data, err := os.Open(jsonF)

	if err != nil {
		panic(err.Error())
	}

	defer func(data *os.File) {
		if err = data.Close(); err != nil {
			panic(err.Error())
		}
	}(data)

	byteData, _ := io.ReadAll(data)

	if err := client.Publish(subject, byteData); err != nil {
		panic(err.Error())
	}
}
