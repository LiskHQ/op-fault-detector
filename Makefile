# makefile

PKGS=$(shell go list ./... | grep -v "/vendor/")

.PHONY: test

test:
	@echo "Test packages"
	@go test -race -shuffle=on -coverprofile=coverage.out -cover $(PKGS)

lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run
format:
	gofmt -s -w .
godocs:
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "open http://localhost:6060/pkg/github.com/LiskHQ/op-fault-detector"
	 godoc -http=:6060