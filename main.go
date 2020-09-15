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
	"github.com/bborbe/teamvault-utils"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/hetzner"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/secret"
	"github.com/bborbe/world/pkg/world"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	_ = pflag.Set("logtostderr", "true")

	ctx, cancel := createContext()
	defer cancel()

	if err := createRootCommand(ctx).Execute(); err != nil {
		if glog.V(4) {
			fmt.Fprintf(os.Stderr, "%+v", err)
		} else {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
		os.Exit(1)
	}
}

func createRootCommand(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "world",
		Short:        "Manage the world",
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().StringP("app", "a", "", "app name")
	rootCmd.PersistentFlags().StringP("cluster", "c", "", "cluster name")
	rootCmd.AddCommand(createApplyCommand(ctx))
	rootCmd.AddCommand(createValidateCommand(ctx))
	rootCmd.AddCommand(createYamlToStructCommand(ctx))
	rootCmd.AddCommand(createSetDnsCommand(ctx))
	return rootCmd
}

func createContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		var canceled bool
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
				if canceled {
					fmt.Println("force exit")
					os.Exit(1)
				}
				fmt.Println("execution canceled")
				cancel()
				canceled = true
			}
		}
	}()
	return ctx, cancel
}

func createApplyCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply the configuration to the world",
		RunE: func(cmd *cobra.Command, args []string) error {
			flag.Parse()
			runner, err := createRunner(ctx, cmd)
			if err != nil {
				return errors.Wrap(err, "create runner failed")
			}
			if err := runner.Validate(ctx); err != nil {
				return errors.Wrap(err, "validate failed")
			}
			glog.V(4).Infof("validate finished")
			if err := runner.Apply(ctx); err != nil {
				return errors.Wrap(err, "apply failed")
			}
			glog.V(4).Infof("apply finished")
			return nil
		},
	}
}

func createValidateCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
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
	}
}
func createSetDnsCommand(ctx context.Context) *cobra.Command {
	command := &cobra.Command{
		Use:   "set-dns",
		Short: "Set dns entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			hostname, err := cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			ip, err := cmd.Flags().GetString("ip")
			if err != nil {
				return err
			}

			dnsSever := &dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: network.Host(hostname),
						IP:   network.IPStatic(ip),
					},
				},
			}

			if err := dnsSever.Apply(ctx); err != nil {
				glog.Fatal(err)
			}
			glog.V(0).Infof("done")
			return nil
		},
	}
	command.Flags().StringP("host", "h", "", "hostname")
	command.Flags().StringP("ip", "i", "", "ip")
	return command
}

func createYamlToStructCommand(context.Context) *cobra.Command {
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
	return command
}

func createRunner(ctx context.Context, cmd *cobra.Command) (*world.Runner, error) {
	teamvaultConfigPath := teamvault.TeamvaultConfigPath("~/.teamvault.json")
	teamvaultConfigPath, err := teamvaultConfigPath.NormalizePath()
	if err != nil {
		return nil, errors.Wrap(err, "normalize teamvaul path failed")
	}
	teamvaultConfig, err := teamvaultConfigPath.Parse()
	if err != nil {
		return nil, errors.Wrap(err, "parse teamvault config failed")
	}
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		glog.V(2).Infof("get parameter app failed: %v", err)
	}
	glog.V(4).Infof("flag app: %s", appName)
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		glog.V(2).Infof("get parameter cluster failed: %v", err)
	}
	glog.V(4).Infof("flag cluster: %s", clusterName)

	builder := world.Builder{
		Configuration: &configuration.World{
			HetznerClient: hetzner.NewClient(),
			App:           configuration.AppName(appName),
			Cluster:       configuration.ClusterName(clusterName),
			TeamvaultSecrets: &secret.Teamvault{
				TeamvaultConnector: teamvault.NewCache(
					teamvault.NewDiskFallbackConnector(
						teamvault.NewRemoteConnector(
							client_builder.New().WithTimeout(5*time.Second).Build().Do,
							teamvaultConfig.Url,
							teamvaultConfig.User,
							teamvaultConfig.Password,
						),
					),
				),
			},
		},
	}
	return builder.Build(ctx)
}
