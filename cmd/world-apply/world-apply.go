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
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("apply all ...")
	if err := run(ctx, &configuration.Configuration{}); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("apply all finished")
}

func run(ctx context.Context, cfg world.Configuration) error {
	glog.V(4).Info("apply configuration ...")
	if cfg.Applier() != nil {
		ok, err := cfg.Applier().Satisfied(ctx)
		if err != nil {
			return err
		}
		if ok {
			glog.V(4).Info("already satisfied => skip")
			return nil
		}
	}
	glog.V(4).Info("found %d childs", len(cfg.Childs()))
	for _, child := range cfg.Childs() {
		if err := run(ctx, child); err != nil {
			return err
		}
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Apply(ctx); err != nil {
			return err
		}
	}
	glog.V(4).Info("apply configuration finished")
	return nil
}
