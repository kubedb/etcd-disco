#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/etcd-manager/lector

source "$REPO_ROOT/hack/libbuild/common/pharmer_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
IMG=lector

DIST=$GOPATH/src/github.com/etcd-manager/lector/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
	export $(cat $DIST/.tag | xargs)
fi

clean() {
    pushd $GOPATH/src/github.com/etcd-manager/lector/hack/docker
    rm lector Dockerfile
    popd
}

build_binary() {
    pushd $GOPATH/src/github.com/etcd-manager/lector
    ./hack/builddeps.sh
    ./hack/make.py build
    detect_tag $DIST/.tag
    popd
}

build_docker() {
    pushd $GOPATH/src/github.com/etcd-manager/lector/hack/docker
    cp $DIST/lector/lector-alpine-amd64 lector
    chmod 755 lector

    cat >Dockerfile <<EOL
FROM ubuntu:16.04

RUN set -x \
  && apt-get update \
  && apt-get install ca-certificates tzdata curl wget openssl -y
RUN wget https://github.com/coreos/etcd/releases/download/v3.3.3/etcd-v3.3.3-linux-amd64.tar.gz -O /tmp/etcd.tar.gz && \
    mkdir /etcd && \
    tar xzvf /tmp/etcd.tar.gz -C /etcd --strip-components=1 && \
    rm /tmp/etcd.tar.gz
RUN cp /etcd/etcdctl /usr/local/bin/

COPY lector /usr/local/bin/etcd



ENTRYPOINT ["etcd etcd"]
EOL
    local cmd="docker build -t pharmer/$IMG:$TAG ."
    echo $cmd; $cmd

    rm lector Dockerfile
    popd
}

build() {
    build_binary
    build_docker
}

docker_push() {
    if [ "$APPSCODE_ENV" = "prod" ]; then
        echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
        exit 0
    fi
    if [ "$TAG_STRATEGY" = "git_tag" ]; then
        echo "Are you trying to 'release' binaries to prod?"
        exit 1
    fi
    hub_canary
}

docker_release() {
    if [ "$APPSCODE_ENV" != "prod" ]; then
        echo "'release' only works in PROD env."
        exit 1
    fi
    if [ "$TAG_STRATEGY" != "git_tag" ]; then
        echo "'apply_tag' to release binaries and/or docker images."
        exit 1
    fi
    hub_up
}

source_repo $@
