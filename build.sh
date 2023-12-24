#! /bin/bash

go build -o output/ferlet cmd/ferlet/main.go
GOOS=linux GOARCH=amd64 go build -o output/server cmd/server/main.go
