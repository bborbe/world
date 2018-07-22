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
	"github.com/pkg/errors"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("building all ...")
	if err := build(ctx, configuration.Apps()); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("building all finished")
}

func build(ctx context.Context, apps world.Apps) error {
	glog.V(1).Infof("build all ...")
	var list []run.RunFunc
	for _, app := range apps {
		list = append(list, func(ctx context.Context) error {
			return run.Sequential(
				ctx,
				app.Builder.Build,
				app.Uploader.Upload,
			)
		})
	}
	glog.V(1).Infof("found %d builds", len(list))
	return errors.Wrap(run.CancelOnFirstError(ctx, list...), "build all failed")
}
