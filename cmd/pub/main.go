package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/broker/stan"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/xeipuuv/gojsonschema"
)

const (
	addr    = "localhost:4222"
	CID     = "clientID"
	clstrID = "wbl0ns"
	jsonF   = "cmd/pub/jsons/test.json"
	jsonFSc = "./api/JSON_schema.json"
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

	if err := validateJSON(byteData, jsonFSc); err != nil {
		logger.L().Error(err.Error())
		return
	}

	if err := client.Publish(subject, byteData); err != nil {
		logger.L().Error(err.Error())
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

func validateJSON(jsonData []byte, schemaPath string) error {
	schemaData, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaData))
	documentLoader := gojsonschema.NewStringLoader(string(jsonData))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var errors string
		for _, desc := range result.Errors() {
			errors += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("JSON не соответствует схеме. Ошибки:\n%s", errors)
	}

	return nil
}
