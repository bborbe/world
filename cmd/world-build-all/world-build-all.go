package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/world/pkg/docker"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	glog.V(1).Infof("building all ...")

	builder := &docker.BuildAll{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := builder.Build(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("building all finished")
}
