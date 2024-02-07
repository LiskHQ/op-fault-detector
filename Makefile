# makefile

PKGS=$(shell go list ./... | grep -v "/vendor/")

.PHONY: test

APP_NAME = faultdetector
GREEN = \033[0;32m
BLUE = \033[0;34m
COLOR_END = \033[0;39m

build:
	@echo "$(BLUE)» Building fault detector application... $(COLOR_END)"
	@CGO_ENABLED=0 go build -v ./...
	@echo "$(GREEN) Binary successfully built$(COLOR_END)"

build-app:
	@echo "$(BLUE)» Building fault detector application binary... $(COLOR_END)"
	@CGO_ENABLED=0 go build -a -o bin/$(APP_NAME) ./cmd/
	@echo "$(GREEN) Binary successfully built$(COLOR_END)"

run-app:
	@./bin/${APP_NAME}

test:
	@echo "Test packages"
	@go test -race -shuffle=on -coverprofile=coverage.out -cover $(PKGS)

test.coverage: test
	go tool cover -func=coverage.out

test.coverage.html: test
	go tool cover -html=coverage.out
	
lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

format:
	gofmt -s -w .

godocs:
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "open http://localhost:6060/pkg/github.com/LiskHQ/op-fault-detector"
	 godoc -http=:6060

.PHONY: docker-build
docker-build:
	@echo "$(BLUE) Building docker image...$(COLOR_END)"
	@docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run:
	@echo "$(BLUE) Running docker image...$(COLOR_END)"
	@docker run -p 8080:8080 $(APP_NAME)