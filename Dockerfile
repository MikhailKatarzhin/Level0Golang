FROM golang:1.21

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
RUN go mod verify

COPY . ./

RUN go build ././cmd/sub/main.go
