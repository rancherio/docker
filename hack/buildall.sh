#!/usr/bin/env bash

cd $(dirname $0)/..

export CGO_ENABLED=1
export KEEPBUNDLE=1

export CC=/usr/bin/arm-linux-gnueabihf-gcc
export DOCKER_CROSSPLATFORMS=linux/arm

hack/make.sh cross

export CC=/usr/bin/aarch64-linux-gnu-gcc
export DOCKER_CROSSPLATFORMS=linux/arm64

hack/make.sh cross
