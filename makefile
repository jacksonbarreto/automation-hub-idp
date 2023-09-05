.PHONY: default run build test doc clean
# Variables
APP_NAME = "IDP"

# Tasks
default: run

run:
	@go run ./make cmd/main.go

build:
	@go build -o $(APP_NAME) ./cmd/main.go

test:
	@go test ./...

docs:
	@swag init -g cmd/main.go

clean:
	@rm -rf $(APP_NAME)
	@rm -rf ./docs