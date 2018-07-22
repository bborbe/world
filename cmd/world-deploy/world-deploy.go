package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	namePtr := flag.String("name", "", "name")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("deploying %s ...", *namePtr)
	app, err := configuration.Apps().WithName(world.Name(*namePtr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
	if err := app.Deployer.Deploy(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "deploy failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("deploying %s finished", *namePtr)
}
