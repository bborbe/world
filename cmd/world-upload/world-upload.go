package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
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

	glog.V(1).Infof("building app %s ...", *namePtr)
	app, err := configuration.Apps().WithName(world.Name(*namePtr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "app %s not found", *namePtr)
		os.Exit(1)
	}
	if err := app.Deployer.GetUploader().GetBuilder().Build(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("building app %s finished", *namePtr)
}
