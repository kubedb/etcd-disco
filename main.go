package main

import (
	"os"

	logs "github.com/appscode/go/log/golog"
	"github.com/etcd-manager/lector/pkg/cmds"
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
