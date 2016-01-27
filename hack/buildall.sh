#!/usr/bin/env bash

cd $(dirname $0)/..

hack/make.sh cross

export CC=/usr/bin/aarch64-linux-gnu-gcc
export DOCKER_CROSSPLATFORMS=linux/arm64
export KEEPBUNDLE=1

hack/make.sh cross
