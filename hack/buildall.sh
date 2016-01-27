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

for d in bundles/latest/cross/linux/*; do
  arch=$(basename $d)
  tardir=$(mktemp -d)
  mkdir ${tardir}/docker
  ls /usr/local/bin/docker-*_${arch} | sed -e 's!^/usr/local/bin/!!' -e "s/_${arch}//" | \
    xargs -I{} cp /usr/local/bin/{}_${arch} ${tardir}/docker/{}
  cp ${d}/$(readlink ${d}/docker) ${tardir}/docker/docker
  tar -czf bundles/$(readlink ${d}/docker)_${arch}.tgz -C ${tardir} docker
done
