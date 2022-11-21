.PHONY: test build build-local update-tracer update-bootstrap-balances \
    run-mainnet-online run-mainnet-offline run-testnet-online run-testnet-offline


GO_PACKAGES=./services/... ./cmd/... ./configuration/... ./wemix/...
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")
TEST_SCRIPT=go test ${GO_PACKAGES}

PWD=$(shell pwd)
NOFILE=100000

test:
	${TEST_SCRIPT}

build:
	docker build -t rosetta-wemix:latest https://github.com/wemixarchive/rosetta-wemix.git

build-local:
	docker build -t rosetta-wemix:latest .

update-tracer:
	curl https://raw.githubusercontent.com/wemixarchive/go-wemix/master/eth/tracers/internal/tracers/call_tracer.js -o wemix/client/call_tracer.js

update-bootstrap-balances:
	go run main.go utils:generate-bootstrap wemix/genesis_files/mainnet.json rosetta-cli-conf/mainnet/bootstrap_balances.json;
	go run main.go utils:generate-bootstrap wemix/genesis_files/testnet.json rosetta-cli-conf/testnet/bootstrap_balances.json;

run-mainnet-online:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/wemix-data:/data" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest

run-mainnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=MAINNET" -e "PORT=8081" -p 8081:8081 rosetta-wemix:latest

run-testnet-online:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -v "${PWD}/wemix-data:/data" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest

run-testnet-offline:
	docker run -d --rm -e "MODE=OFFLINE" -e "NETWORK=TESTNET" -e "PORT=8081" -p 8081:8081 rosetta-wemix:latest

run-mainnet-remote:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "MODE=ONLINE" -e "NETWORK=MAINNET" -e "PORT=8080" -e "GWEMIX=$(gwemix)" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest

run-testnet-remote:
	docker run -d --rm --ulimit "nofile=${NOFILE}:${NOFILE}" -e "MODE=ONLINE" -e "NETWORK=TESTNET" -e "PORT=8080" -e "GWEMIX=$(gwemix)" -p 8080:8080 -p 30303:30303 rosetta-wemix:latest
