# op-fault-detector

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)
![GitHub repo size](https://img.shields.io/github/repo-size/liskhq/op-fault-detector)
![GitHub issues](https://img.shields.io/github/issues-raw/liskhq/op-fault-detector)
![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/liskhq/op-fault-detector)
[![PR CI](https://github.com/LiskHQ/op-fault-detector/actions/workflows/pr.yaml/badge.svg?branch=main&event=merge_group)](https://github.com/LiskHQ/op-fault-detector/actions/workflows/pr.yaml)

Fault detector is a service that identifies mismatches between a local view of the Optimism or superchain network and L2 output proposals published to Ethereum. Here is the reference to the original implementation of the [fault monitoring](https://github.com/ethereum-optimism/optimism/blob/v1.5.0/packages/chain-mon/src/fault-mon/README.md) service from [Optimism](https://www.optimism.io/).

## How it works

The state root of the block is published to the [L2OutputOracle](https://github.com/ethereum-optimism/optimism/blob/39b7262cc3ffd78cd314341b8512b2683c1d9af7/packages/contracts-bedrock/contracts/L1/L2OutputOracle.sol) contract on Ethereum. The `L2OutputOracle` is inferred from the portal contract.

In the application, we take the state root of the given block as reported by an Optimism node, compute `outputRoot` from it and compare it with the `outputRoot` as published to `L2OutputOracle` contract on Ethereum.

## Installation

```
git clone https://github.com/liskhq/op-fault-detector
make build
```

## Running Fault Detector

Copy `config.yaml` file and use any name with `.yaml` extension or edit existing `config.yaml` file to set configuration for the application.

```
make run-app
```
if want to provide custom file, for example, `my-config.yaml`, run,

```
make run-app config=/path/to/my-config.yaml
```

View all available commands by running `make help` and view the commands with options as below.

```sh
build: Builds the application and create a binary at ./bin/

docker-build: Builds docker image

docker-run: Runs docker image, use `make docker-run config={PATH_TO_CONFIG_FILE}` to provide custom config

format: Runs gofmt on the repo

godocs: Runs godoc and serves via endpoint

help: Show help for each of the Makefile recipes

lint: Runs golangci-lint on the repo

run-app: Runs the application, use `make run-app config={PATH_TO_CONFIG_FILE}` to provide custom config

test: Runs tests
```

## Config

The configuration file is used to configure the application. Currently, the default configuration is found under `./config.yaml`. To provide custom config, edit the `./config.yaml` or create own and provide it while running the application `make run-app config={PATH_TO_CUSTOM_CONFIG_FILE}`.

```yaml
# General system configurations
system:
  log_level: "info"

# API related configurations
api:
  server:
    host: "127.0.0.1"
    port: 8080
  base_path: "/api"
  register_versions:
    - v1

# Faultdetector configurations
fault_detector:
  l1_rpc_endpoint: "https://rpc.notadegen.com/eth"
  l2_rpc_endpoint: "https://mainnet.optimism.io/"
  start_batch_index: -1
  l2_output_oracle_contract_address: "0x0000000000000000000000000000000000000000"

```
### System Config
- `system.log_level`: Set log level of the application, by default `info` and available options are `warn`, `debug`, `error` and `fatal`

### API Config
- `api.server.host`: Host of application
- `api.server.port`: Port of application
- `api.base_path`: Base path for the API
- `register_versions`: Versions for APIs

### Fault Detector Config

- `fault_detector.l1_rpc_endpoint`: RPC endpoint for L1 chain.
- `fault_detector.l2_rpc_endpoint`: RPC endpoint for L2 chain.
- `fault_detector.start_batch_index`: Provide batch_index to start from. If not provided, it will pick default `-1` and then application will find the first unfinalized batch index that has not yet passed the fault proof window.
- `fault_detector.l2_output_oracle_contract_address`: Deployed `L2OutputOracle` contract address used to retrieve necessary info for output verification. Only provided for the chains other than Optimism and Lisk Superchain.

## API and Metrics

### API
- Status API exposed via `{api.server.host}:{api.server.port}/api/v1/status`
- Metrics is exposed at `{api.server.host}:{api.server.port}/metrics`
- `{api.server.host}` in `config.yaml` defaults to `127.0.0.1`
- `{api.server.port}` in `config.yaml` defaults to `8080`

### Metrics

```sh
- fault_detector_highest_output_index      prometheus.Gauge     Highest known output index
- fault_detector_is_state_mismatch         prometheus.Gauge     0 if state is ok, 1 if state is mismatched
- fault_detector_api_connection_failure    prometheus.Gauge     Number of API RPC calls failed for L1 and L2 nodes
```
