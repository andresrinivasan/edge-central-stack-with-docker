# TODO: Change Copyright to your company if open sourcing or remove header
#
# Copyright (C) 2021-2022 IOTech Ltd

#build stage
ARG BASE=golang:1.19-alpine3.17
FROM ${BASE} AS builder

ARG ALPINE_PKG_BASE="make git gcc libc-dev libsodium-dev zeromq-dev"

RUN apk add --update --no-cache ${ALPINE_PKG_BASE}
WORKDIR /app

COPY . .

ARG MAKE="make build"
RUN $MAKE

#final stage
FROM alpine:3.17
# TODO: Change Copyright to your company if open sourcing or remove label
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (C) 2021-2022 IOTech Ltd'
LABEL Name=new-app-service Version=${VERSION}

# dumb-init is required as security-bootstrapper uses it in the entrypoint script
RUN apk add --update --no-cache ca-certificates zeromq dumb-init

COPY --from=builder /app/LICENSE /LICENSE
COPY --from=builder /app/res/ /res/
COPY --from=builder /app/new-app-service /new-app-service

# TODO: set this port appropriately as it is in the configuration.toml
EXPOSE 59740

ENTRYPOINT ["/new-app-service"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
