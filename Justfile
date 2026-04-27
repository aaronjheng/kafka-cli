set dotenv-load

bump-deps:
    go get -u ./...
    go mod tidy

lint:
    golangci-lint run --verbose --allow-parallel-runners ./...

lint-with-fix:
    golangci-lint run --verbose --allow-parallel-runners --fix ./...

install:
    go install github.com/aaronjheng/kafka-cli/cmd/kafka@master
