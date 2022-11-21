# Copyright 2020 Coinbase, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Compile golang
FROM ubuntu:20.04 as golang-builder

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app
WORKDIR /app

RUN apt-get update && apt-get install -y curl make gcc g++ git
ENV GOLANG_VERSION 1.19.3
ENV GOLANG_DOWNLOAD_SHA256 74b9640724fd4e6bb0ed2a1bc44ae813a03f1e72a4c76253e2d5c015494430ba
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Compile gwemix
FROM golang-builder as gwemix-builder

# VERSION: go-wemix w0.10.1
RUN git clone https://github.com/wemixarchive/go-wemix \
  && cd go-wemix \
  && git checkout master

RUN cd go-wemix \
  && make USE_ROCKSDB=NO

RUN mv go-wemix/build/bin/gwemix /app/gwemix \
  && rm -rf go-wemix

# Compile rosetta-wemix
FROM golang-builder as rosetta-builder

# Use native remote build context to build in any directory
COPY . src
RUN cd src \
  && go build

RUN mv src/rosetta-wemix /app/rosetta-wemix \
  && mkdir /app/wemix \
  && mv src/wemix/call_tracer.js /app/wemix/call_tracer.js \
  && mv src/wemix/gwemix.toml /app/wemix/gwemix.toml \
  && rm -rf src

## Build Final Image
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /app \
  && chown -R nobody:nogroup /app \
  && mkdir -p /data \
  && chown -R nobody:nogroup /data

WORKDIR /app

# Copy binary from gwemix-builder
COPY --from=gwemix-builder /app/gwemix /app/gwemix

# Copy binary from rosetta-builder
COPY --from=rosetta-builder /app/wemix /app/wemix
COPY --from=rosetta-builder /app/rosetta-wemix /app/rosetta-wemix

# Set permissions for everything added to /app
RUN chmod -R 755 /app/*

CMD ["/app/rosetta-wemix", "run"]
