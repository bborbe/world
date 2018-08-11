package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/model"
	"github.com/bborbe/world/configuration"
	"github.com/golang/glog"
	"github.com/bborbe/world"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("validate all ...")

	teamvaultConfigPath := model.TeamvaultConfigPath("~/.teamvault.json")
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
	conf := &configuration.Configuration{
		TeamvaultConnector: connector.New(httpClient.Do, teamvaultConfig.Url, teamvaultConfig.User, teamvaultConfig.Password),
	}

	if err := world.Validate(ctx, conf); err != nil {
		fmt.Fprintf(os.Stderr, "validate failed: %+v\n", err)
		os.Exit(1)
	}

	glog.V(1).Infof("validate all finished")
}
