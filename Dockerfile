FROM golang:1.18 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

COPY main.go main.go
COPY fp_object.go fp_object.go
RUN go mod tidy


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o fo main.go fp_object.go