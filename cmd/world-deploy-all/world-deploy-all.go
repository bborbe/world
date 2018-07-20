package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	glog.V(1).Infof("deploying all ...")

	deployer := &k8s.DeployAll{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := deployer.Deploy(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "deploy failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("deploying all finished")
}
