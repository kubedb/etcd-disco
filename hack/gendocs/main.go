package main

import (
	"fmt"
	"log"
	"os"
	"github.com/appscode/etcd-disco/pkg/cmds"
	"github.com/appscode/go/runtime"
	"github.com/spf13/cobra/doc"
)

// ref: https://github.com/spf13/cobra/blob/master/doc/md_docs.md
func main() {
	rootCmd := cmds.NewCmdCluster()
	dir := runtime.GOPath() + "/src/github.com/appscode/etcd-disco/docs/reference"
	fmt.Printf("Generating cli markdown tree in: %v\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	doc.GenMarkdownTree(rootCmd, dir)
}
