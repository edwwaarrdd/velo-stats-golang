.PHONY: build test fmt vet run worker up down

build:
	go build -o velo ./cmd/velo

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

run:
	go run ./cmd/velo serve

worker:
	go run ./cmd/velo work -queue=default

up:
	docker compose up -d --build

down:
	docker compose down
