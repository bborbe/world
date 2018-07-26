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
	applier := &Applier{
		Apps: configuration.Apps(),
	}
	if err := applier.Apply(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("apply all finished")
}

type Applier struct {
	Apps world.Apps
}

func (a *Applier) Apply(ctx context.Context) error {
	var list []run.RunFunc
	for _, app := range a.Apps {
		list = append(list, buildFunc(app))
	}
	return run.All(ctx, list...)
}

func buildFunc(app world.App) run.RunFunc {
	return func(ctx context.Context) error {
		glog.V(4).Infof("apply app ...")
		defer glog.V(4).Infof("apply app finished")
		return run.Sequential(
			ctx,
			app.Validate,
			func(ctx context.Context) error {
				ok, err := app.Builder().Satisfied(ctx)
				if err != nil {
					return err
				}
				if ok {
					return nil
				}
				return app.Builder().Build(ctx)
			},
			func(ctx context.Context) error {
				ok, err := app.Uploader().Satisfied(ctx)
				if err != nil {
					return err
				}
				if ok {
					return nil
				}
				return app.Uploader().Upload(ctx)
			},
			func(ctx context.Context) error {
				ok, err := app.Deployer().Satisfied(ctx)
				if err != nil {
					return err
				}
				if ok {
					return nil
				}
				return app.Deployer().Deploy(ctx)
			},
		)
	}
}
