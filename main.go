package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/bborbe/http/client_builder"
	"github.com/bborbe/teamvault-utils"
	teamvaultconnector "github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/world"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Set("v", "2")
	flag.Set("logtostderr", "true")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd := &cobra.Command{
		Use:   "world",
		Short: "Manage the world",
	}
	rootCmd.PersistentFlags().StringP("app", "a", "", "app name")
	rootCmd.PersistentFlags().StringP("cluster", "c", "", "cluster name")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "apply",
		Short: "Apply the configuration to the world",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner, err := createRunner(ctx, cmd)
			if err != nil {
				return err
			}
			if err := runner.Validate(ctx); err != nil {
				return err
			}
			if err := runner.Apply(ctx); err != nil {
				return err
			}
			return nil
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration of the world",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner, err := createRunner(ctx, cmd)
			if err != nil {
				return err
			}
			if err := runner.Validate(ctx); err != nil {
				return err
			}
			return nil
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func createRunner(ctx context.Context, cmd *cobra.Command) (*world.Runner, error) {

	teamvaultConfigPath := teamvault.TeamvaultConfigPath("~/.teamvault.json")
	teamvaultConfigPath, err := teamvaultConfigPath.NormalizePath()
	if err != nil {
		return nil, err
	}
	teamvaultConfig, err := teamvaultConfigPath.Parse()
	if err != nil {
		return nil, err
	}

	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		glog.V(2).Infof("get parameter app failed: %v", err)
	}
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		glog.V(2).Infof("get parameter cluster failed: %v", err)
	}

	builder := world.Builder{
		Configuration: &configuration.World{
			App:     configuration.AppName(appName),
			Cluster: configuration.ClusterName(clusterName),
			TeamvaultSecrets: &secret.Teamvault{
				TeamvaultConnector: teamvaultconnector.NewCache(
					&teamvaultconnector.DiskFallback{
						Connector: teamvaultconnector.NewRemote(client_builder.New().WithTimeout(5*time.Second).Build().Do, teamvaultConfig.Url, teamvaultConfig.User, teamvaultConfig.Password),
					},
				),
			},
		},
	}
	return builder.Build(ctx)
}
