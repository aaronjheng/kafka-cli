set dotenv-load

lint:
	golangci-lint run --allow-parallel-runners
