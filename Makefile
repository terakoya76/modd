go.lint:
	# for local
	# GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0; go mod tidy
	@if [ -z `which golangci-lint 2> /dev/null` ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $(go env GOPATH)/bin v1.43.0; \
	fi
	gofmt -s -w .
	golangci-lint run --timeout 2m

go.test: go.lint
	go test -race -v ./...
