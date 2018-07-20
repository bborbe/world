package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	namePtr := flag.String("name", "", "name")
	flag.Parse()

	glog.V(1).Infof("building app %s ...", *namePtr)

	app, err := world.GetApp(world.Name(*namePtr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "app %s not found", *namePtr)
		os.Exit(1)
	}
	builder, err := docker.BuilderForApp(*app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "builder for app %s not found", *namePtr)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := builder.Build(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("building app %s finished", *namePtr)
}
