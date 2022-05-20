go.lint:
	gofmt -s -w .
	golangci-lint run --timeout 2m

go.test: go.lint
	go test -race -v ./...
