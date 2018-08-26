package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault-utils"
	teamvaultconnector "github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/runner"
	"github.com/bborbe/world/pkg/secret"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	teamvaultConfigPath := teamvault.TeamvaultConfigPath("~/.teamvault.json")
	teamvaultConfigPath, err := teamvaultConfigPath.NormalizePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "normalize teamvault config path failed: %+v\n", err)
		os.Exit(1)
	}
	teamvaultConfig, err := teamvaultConfigPath.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse teamvault config failed: %+v\n", err)
		os.Exit(1)
	}

	httpClient := client_builder.New().WithTimeout(5 * time.Second).Build()
	world := &configuration.World{
		TeamvaultSecrets: &secret.Teamvault{
			TeamvaultConnector: teamvaultconnector.NewCache(
				&teamvaultconnector.DiskFallback{
					Connector: teamvaultconnector.NewRemote(httpClient.Do, teamvaultConfig.Url, teamvaultConfig.User, teamvaultConfig.Password),
				},
			),
		},
	}
	flag.StringVar(&world.Name, "name", "", "app name")
	flag.Parse()

	conf, err := world.Configuration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find configuration for %s failed: %v\n", world.Name, err)
		os.Exit(1)
	}
	builder := runner.Builder{
		Configuration: conf,
	}
	r, err := builder.Build(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create runtime failed: %+v\n", err)
		os.Exit(1)
	}

	glog.V(1).Infof("validate all ...")
	if err := r.Validate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "validate failed: %+v\n", err)
		os.Exit(1)
	}
	glog.V(1).Infof("validate all finished")
}
