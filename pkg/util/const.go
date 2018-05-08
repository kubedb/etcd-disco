package util

import "time"

const (
	DefaultStartTimeout          = 900 * time.Second
	DefaultStartRejoinTimeout    = 60 * time.Second
	DefaultMemberCleanerInterval = 15 * time.Second
)
const (
	DefaultClientPort  = 2379
	DefaultPeerPort    = 2380
	DefaultMetricsPort = 2381

	DefaultDialTimeout    = 5 * time.Second
	DefaultRequestTimeout = 5 * time.Second
	DefaultAutoSync       = 1 * time.Second
)
