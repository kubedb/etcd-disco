package main

import (
	"os"

	"github.com/appscode/etcd-disco/pkg/cmds"
	logs "github.com/appscode/go/log/golog"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	/*if err := cmds.NewRootCmd(os.Stdin, os.Stdout, os.Stderr, Version).Execute(); err != nil {
		os.Exit(1)
	}*/
	cmds.NewCmdCluster().Execute()
	os.Exit(0)
}
