#!/bin/sh

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o eli-serv main.go