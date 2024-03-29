# makefile

PKGS=$(shell go list ./... | grep -v "/vendor/")

APP_NAME = faultdetector
GREEN = \033[1;32m
BLUE = \033[1;34m
COLOR_END = \033[0;39m
CONFIG_PATH = $(realpath $(config))

build: # Builds the application and create a binary at ./bin/
	@echo "$(BLUE)» Building fault detector application binary... $(COLOR_END)"
	@CGO_ENABLED=0 go build -a -o bin/$(APP_NAME) ./cmd/...
	@echo "$(GREEN) Binary successfully built$(COLOR_END)"

install: # Installs faultdetector cmd and creates executable at $GOPATH/bin/
	@echo "$(BLUE)» Installing fault detector command... $(COLOR_END)"
	@CGO_ENABLED=0 go install ./cmd/$(APP_NAME)
	@echo "$(GREEN) $(APP_NAME) successfully installed$(COLOR_END)"

run-app: # Runs the application, use `make run-app config={PATH_TO_CONFIG_FILE}` to provide custom config
ifdef config
	@./bin/${APP_NAME} --config $(config)
else
	@./bin/${APP_NAME}
endif

.PHONY: test
test: # Runs tests
	@echo "Test packages"
	@go test -race -shuffle=on -coverprofile=coverage.out -cover $(PKGS)

test.coverage: test
	go tool cover -func=coverage.out

test.coverage.html: test
	go tool cover -html=coverage.out

test.e2e: # Runs e2e tests
	@echo "Test e2e"
	@go test -v ./cmd/$(APP_NAME)
	
lint: # Runs golangci-lint on the repo
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

format: # Runs gofmt on the repo
	gofmt -s -w .

godocs: # Runs godoc and serves via endpoint
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "open http://localhost:6060/pkg/github.com/LiskHQ/op-fault-detector"
	 godoc -http=:6060

.PHONY: docker-build
docker-build: # Builds docker image
	@echo "$(BLUE) Building docker image...$(COLOR_END)"
	@docker build -t $(APP_NAME) .

.PHONY: docker-run
docker-run: # Runs docker image, use `make docker-run config={PATH_TO_CONFIG_FILE}` to provide custom config and to provide slack access token use `make docker-run slack_access_token={ACCESS_TOKEN}`
	@echo "$(BLUE) Starting docker container $(APP_NAME)...$(COLOR_END)"
ifneq ($(and $(config),$(slack_access_token)),)
	@docker run --name $(APP_NAME) -p 8080:8080 -d -v $(CONFIG_PATH):/home/onchain/faultdetector/config.yaml -t -e SLACK_ACCESS_TOKEN_KEY=$(slack_access_token) $(APP_NAME)
else ifdef config
	@docker run --name $(APP_NAME) -p 8080:8080 -d -v $(CONFIG_PATH):/home/onchain/faultdetector/config.yaml -t $(APP_NAME)
else ifdef slack_access_token
	@docker run --name $(APP_NAME) -p 8080:8080 -d -t -e SLACK_ACCESS_TOKEN_KEY=$(slack_access_token) $(APP_NAME)
else
	@docker run --name $(APP_NAME) -p 8080:8080 -d $(APP_NAME)
endif

docker-stop:
	@echo "$(BLUE) Stopping and removing docker container $(APP_NAME)...$(COLOR_END)"
	@docker rm -f $(APP_NAME)

.PHONY: help
help: # Show help for each of the Makefile recipes
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "$(GREEN)$$(echo $$l | cut -f 1 -d':')$(COLOR_END):$$(echo $$l | cut -f 2- -d'#')\n"; done
