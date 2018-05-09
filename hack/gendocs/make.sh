#!/usr/bin/env bash

pushd $GOPATH/src/github.com/etcd-manager/lector/hack/gendocs
go run main.go
popd
