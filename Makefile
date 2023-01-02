-include .env

.ONESHELL:
.SHELLFLAGS = -ec

.PHONY: lint
lint:
	@golangci-lint run --allow-parallel-runners
