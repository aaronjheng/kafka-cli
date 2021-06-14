include .bingo/Variables.mk

.ONESHELL:

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run --allow-parallel-runners
