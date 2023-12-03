package main

import (
	"fmt"
	"io"
	"os"
	"strings"

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

	byteData, err := readJsonFromFile()
	if err != nil {
		panic(err.Error())
	}

	//TODO check JSON with JSON schema

	if err := client.Publish(subject, byteData); err != nil {
		panic(err.Error())
	}
}

func readJsonFromFile() ([]byte, error) {
	file, err := os.Open(jsonF)

	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			panic(err.Error())
		}
	}(file)

	byteData, _ := io.ReadAll(file)

	fileString := string(byteData[:])

	fileString = TrimJsonFileString(fileString)

	return []byte(fileString), nil
}

func TrimJsonFileString(jsonFileString string) string {

	trimedString := strings.ReplaceAll(jsonFileString, "\r", "")
	trimedString = strings.ReplaceAll(trimedString, "\n", "")

	//Non-necessary for unmarshal, but it shorted string -> less length -> compact upload
	trimedString = strings.ReplaceAll(trimedString, "      ", " ")
	trimedString = strings.ReplaceAll(trimedString, "    ", " ")
	trimedString = strings.ReplaceAll(trimedString, "  ", " ")

	return trimedString
}
