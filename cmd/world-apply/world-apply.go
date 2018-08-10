package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"time"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/run"
	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/model"
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

	teamvaultConfigPath := model.TeamvaultConfigPath("~/.teamvault.json")
	teamvaultConfigPath, err := teamvaultConfigPath.NormalizePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "normalize teamvault config path failed: %v", err)
		os.Exit(1)
	}
	teamvaultConfig, err := teamvaultConfigPath.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse teamvault config failed: %v", err)
		os.Exit(1)
	}

	httpClient := client_builder.New().WithTimeout(5 * time.Second).Build()
	conf := &configuration.Configuration{
		TeamvaultConnector: connector.New(httpClient.Do, teamvaultConfig.Url, teamvaultConfig.User, teamvaultConfig.Password),
	}

	if err := validate(ctx, conf); err != nil {
		fmt.Fprintf(os.Stderr, "validate failed: %v", err)
		os.Exit(1)
	}

	if err := apply(ctx, conf); err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("apply all finished")
}

func validate(ctx context.Context, cfg world.Configuration) error {
	if err := cfg.Validate(ctx); err != nil {
		return err
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Validate(ctx); err != nil {
			return err
		}
	}
	for _, child := range cfg.Childs() {
		if err := validate(ctx, child); err != nil {
			return err
		}
	}
	return nil
}

func apply(ctx context.Context, cfg world.Configuration) error {
	glog.V(4).Infof("apply configuration ...")
	if cfg.Applier() != nil {
		ok, err := cfg.Applier().Satisfied(ctx)
		if err != nil {
			return err
		}
		if ok {
			glog.V(4).Infof("already satisfied => skip")
			return nil
		}
	}
	glog.V(4).Infof("found %d childs", len(cfg.Childs()))

	var list []run.RunFunc
	for _, child := range cfg.Childs() {
		list = append(list, func(child world.Configuration) run.RunFunc {
			return func(ctx context.Context) error {
				return apply(ctx, child)
			}
		}(child))
	}
	if err := run.Sequential(ctx, list...); err != nil {
		return err
	}
	if cfg.Applier() != nil {
		if err := cfg.Applier().Apply(ctx); err != nil {
			return err
		}
	}
	glog.V(4).Infof("apply configuration finished")
	return nil
}
