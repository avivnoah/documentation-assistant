SHELL := /bin/bash

BINARY := documentation-assistant
BIN_DIR := bin
PY_SCRIPT := app/core.py
ENV_FILE := .env

.PHONY: all build run-go run-python run-both clean

all: build

build:
	mkdir -p $(BIN_DIR)
	go mod tidy
	go build -o $(BIN_DIR)/$(BINARY) .

run-go: build
	@if [ -f $(ENV_FILE) ]; then set -a; . $(ENV_FILE); set +a; fi; \
	$(BIN_DIR)/$(BINARY)

run-python:
	@if [ -f $(ENV_FILE) ]; then set -a; . $(ENV_FILE); set +a; fi; \
	PYTHONUNBUFFERED=1 streamlit run $(PY_SCRIPT)

clean:
	rm -rf $(BIN_DIR)