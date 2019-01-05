// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bborbe/http/client_builder"
	teamvault "github.com/bborbe/teamvault-utils"
	teamvaultconnector "github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	_ = pflag.Set("logtostderr", "true")

	rootCmd := &cobra.Command{
		Use:          "world",
		Short:        "Manage the world",
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().StringP("app", "a", "", "app name")
	rootCmd.PersistentFlags().StringP("cluster", "c", "", "cluster name")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "apply",
		Short: "Apply the configuration to the world",
		RunE: func(cmd *cobra.Command, args []string) error {
			flag.Parse()
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
	command := &cobra.Command{
		Use:   "yaml-to-struct",
		Short: "Convert the given yaml to world struct",
		RunE: func(cmd *cobra.Command, args []string) error {
			filename, err := cmd.Flags().GetString("file")
			if err != nil {
				return errors.Wrap(err, "get parameter file failed")
			}
			file, err := os.Open(filename)
			if err != nil {
				return errors.Wrap(err, "open file failed")
			}
			defer file.Close()
			return errors.Wrap(k8s.YamlToStruct(file, os.Stdout), "convert to struct failed")
		},
	}
	command.Flags().StringP("file", "f", "", "filename")
	rootCmd.AddCommand(command)

	if err := rootCmd.Execute(); err != nil {
		if glog.V(4) {
			fmt.Fprintf(os.Stderr, "%+v", err)
		} else {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
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
