PROJECT_NAME=antibot-developer-trainee

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o bin/main ./cmd/main.go

docker-build:
	docker build -t ${PROJECT_NAME} .

up:
	docker-compose up

down:
	docker-compose down

run:
	go run ./cmd/main.go

test-coverage:
	go test -v -coverprofile=./report/coverage.out -cover `go list ./... | grep -v mocks`
	go tool cover -func=./report/coverage.out
	go tool cover -html=./report/coverage.out

.DEFAULT_GOAL := build