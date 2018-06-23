#!/usr/bin/env bash

etcd-disco cluster create --name infra1 \
      --initial-advertise-peer-urls http://127.0.0.1:2380 \
      --listen-peer-urls http://127.0.0.1:2380 \
      --listen-client-urls http://127.0.0.1:2379 \
      --advertise-client-urls http://127.0.0.1:2379


etcd-disco cluster join  --name infra2 \
      --initial-advertise-peer-urls http://127.0.0.2:2380 \
      --listen-peer-urls http://127.0.0.2:2380 \
      --listen-client-urls http://127.0.0.2:2379 \
      --advertise-client-urls http://127.0.0.2:2379 \
      --server-address=http://127.0.0.1:2379
