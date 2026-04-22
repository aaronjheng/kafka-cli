set dotenv-load := true

bump-deps:
    go get -u ./...
    go mod tidy

lint:
    golangci-lint run --verbose --allow-parallel-runners

lint-with-fix:
    golangci-lint run --verbose --allow-parallel-runners --fix
