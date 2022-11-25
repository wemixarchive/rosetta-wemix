<p align="center">
  <a href="https://www.rosetta-api.org">
    <img width="90%" alt="Rosetta" src="https://www.rosetta-api.org/img/rosetta_header.png">
  </a>
</p>
<h3 align="center">
   Rosetta Wemix
</h3>
<p align="center"><b>
ROSETTA-WEMIX IS CONSIDERED <a href="https://en.wikipedia.org/wiki/Software_release_life_cycle#Alpha">ALPHA SOFTWARE</a>.
USE AT YOUR OWN RISK!
</b></p>

## Overview
`rosetta-wemix` provides a reference implementation of the Rosetta API for Wemix in Golang. If you haven't heard of the Rosetta API, you can find more information [here](https://rosetta-api.org).

## Features
* Comprehensive tracking of all WEMIX balance changes
* Atomic balance lookups using go-ethereum's GraphQL Endpoint
* Idempotent access to all transaction traces and receipts

## System Requirements
`rosetta-wemix` has been tested on an [AWS c5.2xlarge instance](https://aws.amazon.com/ec2/instance-types/c5).
This instance type has 8 vCPU and 16 GB of RAM. If you use a computer with less than 16 GB of RAM,
it is possible that `rosetta-wemix` will exit with an OOM error.

### Recommended OS Settings
To increase the load `rosetta-wemix` can handle, it is recommended to tune your OS
settings to allow for more connections. On a linux-based OS, you can run the following
commands ([source](http://www.tweaked.io/guide/kernel)):
```text
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
sysctl -w net.ipv4.tcp_max_syn_backlog=10000
sysctl -w net.core.somaxconn=10000
sysctl -p (when done)
```
_We have not tested `rosetta-wemix` with `net.ipv4.tcp_tw_recycle` and do not recommend
enabling it._

You should also modify your open file settings to `100000`. This can be done on a linux-based OS
with the command: `ulimit -n 100000`.

## Usage
As specified in the [Rosetta API Principles](https://www.rosetta-api.org/docs/automated_deployment.html),
all Rosetta implementations must be deployable via Docker and support running via either an
[`online` or `offline` mode](https://www.rosetta-api.org/docs/node_deployment.html#multiple-modes).

**YOU MUST INSTALL DOCKER FOR THE FOLLOWING INSTRUCTIONS TO WORK. YOU CAN DOWNLOAD
DOCKER [HERE](https://www.docker.com/get-started).**

### Install
Running the following commands will create a Docker image called `rosetta-wemix:latest`.

#### From Source
After cloning this repository, run:
```text
make build-local
```

### Run
Running the following commands will start a Docker container in
[detached mode](https://docs.docker.com/engine/reference/run/#detached--d) with
a data directory at `<working directory>/wemix-data` and the Rosetta API accessible
at port `8080`.

#### Configuration Environment Variables
* `MODE` (required) - Determines if Rosetta can make outbound connections. Options: `ONLINE` or `OFFLINE`.
* `NETWORK` (required) - Ethereum network to launch and/or communicate with. Options: `MAINNET` or `TESTNET` (which defaults to `TESTNET` for backwards compatibility).
* `PORT`(required) - Which port to use for Rosetta.
* `GWEMIX` (optional) - Point to a remote `gwemix` node instead of initializing one
* `SKIP_GWEMIX_ADMIN` (optional, default: `FALSE`) - Instruct Rosetta to not use the `gwemix` `admin` RPC calls. This is typically disabled by hosted blockchain node services.

#### Mainnet:Online
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/wemix-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-mainnet-online`._

#### Mainnet:Online (Remote)
```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "GWEMIX=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-mainnet-remote gwemix=<NODE URL>`._

#### Mainnet:Offline
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-mainnet-offline`._

#### Testnet:Online
```text
docker run -d --rm --ulimit "nofile=100000:100000" -v "$(pwd)/wemix-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-testnet-online`._

#### Testnet:Online (Remote)
```text
docker run -d --rm --ulimit "nofile=100000:100000" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "GWEMIX=<NODE URL>" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-testnet-remote gwemix=<NODE URL>`._

#### Testnet:Offline
```text
docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-wemix:latest
```
_If you cloned the repository, you can run `make run-testnet-offline`._

## License
This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).

© 2022 Wemix