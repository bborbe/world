package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("apply all ...")
	applierAll := &ApplierAll{
		Apps: configuration.Apps(),
	}
	if err := applierAll.Apply(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("apply all finished")
}

type ApplierAll struct {
	Apps world.Apps
}

func (a *ApplierAll) Apply(ctx context.Context) error {
	var list []run.RunFunc
	for _, app := range a.Apps {
		list = append(list, app.Apply)
	}
	return run.All(ctx, list...)
}
