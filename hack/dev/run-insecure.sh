etcd --name infra1 \
  --initial-advertise-peer-urls http://127.0.0.1:2380 \
  --listen-peer-urls http://127.0.0.1:2380 \
  --listen-client-urls http://127.0.0.1:2379 \
  --advertise-client-urls http://127.0.0.1:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster infra1=http://127.0.0.1:2380 \
  --initial-cluster-state new \
  --data-dir=/tmp/infra1

# ---------------------------------------------------------------------
etcdctl member add infra2 http://127.0.0.2:2380
Added member named infra2 with ID 9ac1409d87f19f5f to cluster

ETCD_NAME="infra2"
ETCD_INITIAL_CLUSTER="infra1=http://127.0.0.1:2380,infra2=http://127.0.0.2:2380"
ETCD_INITIAL_CLUSTER_STATE="existing"

etcd --name infra2 \
  --initial-advertise-peer-urls http://127.0.0.2:2380 \
  --listen-peer-urls http://127.0.0.2:2380 \
  --listen-client-urls http://127.0.0.2:2379 \
  --advertise-client-urls http://127.0.0.2:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster infra1=http://127.0.0.1:2380,infra2=http://127.0.0.2:2380 \
  --initial-cluster-state existing \
  --data-dir=/tmp/infra2


$ etcdctl member list
9ac1409d87f19f5f: name=infra2 peerURLs=http://127.0.0.2:2380 clientURLs=http://127.0.0.2:2379 isLeader=false
bf9071f4639c75cc: name=infra1 peerURLs=http://127.0.0.1:2380 clientURLs=http://127.0.0.1:2379 isLeader=true

# ---------------------------------------------------------------------
$ etcdctl member add infra3 http://127.0.0.3:2380
Added member named infra3 with ID 4de2f3ec1d4961c to cluster

ETCD_NAME="infra3"
ETCD_INITIAL_CLUSTER="infra1=http://127.0.0.1:2380,infra2=http://127.0.0.2:2380,infra3=http://127.0.0.3:2380"
ETCD_INITIAL_CLUSTER_STATE="existing"

etcd --name infra3 \
  --initial-advertise-peer-urls http://127.0.0.3:2380 \
  --listen-peer-urls http://127.0.0.3:2380 \
  --listen-client-urls http://127.0.0.3:2379 \
  --advertise-client-urls http://127.0.0.3:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster infra1=http://127.0.0.1:2380,infra2=http://127.0.0.2:2380,infra3=http://127.0.0.3:2380 \
  --initial-cluster-state existing \
  --data-dir=/tmp/infra3

$ etcdctl member list
4de2f3ec1d4961c: name=infra3 peerURLs=http://127.0.0.3:2380 clientURLs=http://127.0.0.3:2379 isLeader=false
9ac1409d87f19f5f: name=infra2 peerURLs=http://127.0.0.2:2380 clientURLs=http://127.0.0.2:2379 isLeader=false
bf9071f4639c75cc: name=infra1 peerURLs=http://127.0.0.1:2380 clientURLs=http://127.0.0.1:2379 isLeader=true
