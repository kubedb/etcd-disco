#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/etcd-disco/hack/gendocs
go run main.go
popd
