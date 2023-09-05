.PHONY: default run build test doc clean
# Variables
APP_NAME = "IDP"

# Determine the platform
ifdef ComSpec
    RM = if exist "$(APP_NAME)$(APP_EXT)" del /Q /F
    RMDIR = if exist
    APP_EXT = .exe
else
    RM = rm -rf
    RMDIR = rm -rf
    APP_EXT =
endif

# Tasks
default: run

run:
	@go run ./cmd/main.go

build:
	@go build -o $(APP_NAME) ./cmd/main.go

test:
	@go test ./...

docs:
	@swag init -g cmd/main.go

clean:
	@$(RM) "$(APP_NAME)$(APP_EXT)"
	@if exist docs rmdir /S /Q docs