package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	namePtr := flag.String("name", "", "name")
	flag.Parse()

	glog.V(1).Infof("deploying %s ...", *namePtr)

	app, err := world.GetApp(world.Name(*namePtr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	deployer, err := k8s.DeployerForApp(*app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := deployer.Deploy(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "deploy failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("deploying %s finished", *namePtr)
}
