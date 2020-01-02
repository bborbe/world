// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

func NewBackupConfigJson(targets BackupTargets) BackupConfigJson {
	var result BackupConfigJson
	for _, target := range targets {
		result = append(result, BackupConfigJsonEntry{
			Active:      true,
			User:        target.User,
			Host:        target.Host,
			IP:          target.IP,
			Port:        target.Port,
			ExcludeFrom: fmt.Sprintf("/config/%s.exclude", target.Host),
			Directory:   target.Directory,
		})
	}
	return result
}

type BackupConfigJson []BackupConfigJsonEntry

type BackupConfigJsonEntry struct {
	Active      bool   `json:"active"`
	User        string `json:"user"`
	Host        string `json:"host"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	ExcludeFrom string `json:"exclude_from"`
	Directory   string `json:"dir"`
}

func (b BackupConfigJson) Value(ctx context.Context) (string, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(b); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (b BackupConfigJson) Validate(ctx context.Context) error {
	return json.NewEncoder(&bytes.Buffer{}).Encode(b)
}

type BackupTargets []BackupTarget

func (b BackupTargets) Validate(ctx context.Context) error {
	for _, target := range b {
		if err := target.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type BackupTarget struct {
	User      string
	Host      string
	IP        string
	Port      int
	Excludes  []string
	Directory string
}

func (b BackupTarget) Validate(ctx context.Context) error {
	if b.Directory == "" {
		return errors.New("Directory missing")
	}
	if b.User == "" {
		return errors.New("User missing")
	}
	if b.Host == "" {
		return errors.New("Host missing")
	}
	if b.Port <= 0 || b.Port >= 65535 {
		return errors.Errorf("invalid port %d", b.Port)
	}
	return nil
}

type BackupClient struct {
	Context       k8s.Context
	Domains       k8s.IngressHosts
	BackupSshKey  deployer.SecretValue
	BackupTargets BackupTargets
}

func (b *BackupClient) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		b.Context,
		b.Domains,
		b.BackupSshKey,
		b.BackupTargets,
	)
}

func (b *BackupClient) Children() []world.Configuration {
	result := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: b.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "backup",
					Name:      "backup",
				},
			},
		},
	}
	result = append(result, b.rsync()...)
	result = append(result, b.cleanup()...)
	result = append(result, b.statusServer()...)
	result = append(result, b.statusClient()...)
	return result
}

func (b *BackupClient) rsync() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-client",
		Tag:        "1.1.0",
	}
	configValues := deployer.ConfigValues{
		"backup-config.json": NewBackupConfigJson(b.BackupTargets),
	}
	for _, target := range b.BackupTargets {
		buf := &bytes.Buffer{}
		sort.Strings(target.Excludes)
		for _, exclude := range target.Excludes {
			exclude = strings.ReplaceAll(exclude, ` `, `\ `)
			fmt.Fprintf(buf, "- %s\n", exclude)
		}
		configValues[fmt.Sprintf("%s.exclude", target.Host)] = deployer.ConfigValueStatic(buf.String())
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.ConfigMapApplier{
				Context:      b.Context,
				Namespace:    "backup",
				Name:         "config",
				ConfigValues: configValues,
			},
		),
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   b.Context,
				Namespace: "backup",
				Name:      "backup",
				Secrets: deployer.Secrets{
					"backup-ssh-key": b.BackupSshKey,
				},
			},
		),
		&deployer.DeploymentDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "rsync",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Annotations: map[string]string{
				"config-checksum": configValues.Checksum(),
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "rsync",
					Image: image,
					Requirement: &build.BackupRsyncClient{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "2000m",
							Memory: "1000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "500Mi",
						},
					},
					Args: []k8s.Arg{
						"-logtostderr",
						"-v=4",
					},
					Env: []k8s.Env{
						{
							Name:  "CONFIG",
							Value: "/config/backup-config.json",
						},
						{
							Name:  "TARGET",
							Value: "/backup",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/backup-rsync-client.lock",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "config",
							Path:     "/config",
							ReadOnly: true,
						},
						{
							Name:     "ssh",
							Path:     "/root/.ssh",
							ReadOnly: true,
						},
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "config",
					ConfigMap: k8s.PodVolumeConfigMap{
						Name: "config",
					},
				},
				{
					Name: "backup",
					Host: k8s.PodVolumeHost{
						Path: "/backup",
					},
				},
				{
					Name: "ssh",
					Secret: k8s.PodVolumeSecret{
						Name: "backup",
						Items: []k8s.PodSecretItem{
							{
								Key:  "backup-ssh-key",
								Mode: 384,
								Path: "id_rsa",
							},
						},
					},
				},
			},
		},
	}
}

func (b *BackupClient) cleanup() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/backup-rsync-cleanup",
		Tag:        "2.0.0",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "cleanup",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "cleanup",
					Image: image,
					Requirement: &build.BackupRsyncCleanup{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args: []k8s.Arg{"-logtostderr", "-v=1"},
					Env: []k8s.Env{
						{
							Name:  "TARGET",
							Value: "/backup",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/backup-rsync-cleanup.lock",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "backup",
							Path: "/backup",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Host: k8s.PodVolumeHost{
						Path: "/backup",
					},
				},
			},
		},
	}
}

func (b *BackupClient) statusServer() []world.Configuration {
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	image := docker.Image{
		Repository: "bborbe/backup-status-server",
		Tag:        "2.0.0",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "status-server",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "backup",
					Image: image,
					Requirement: &build.BackupStatusServer{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
						{
							Name:  "ROOTDIR",
							Value: "/backup",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "backup",
							Path:     "/backup",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "backup",
					Host: k8s.PodVolumeHost{
						Path: "/backup",
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "status-server",
			Ports:     []deployer.Port{port},
		},
	}
}

func (b *BackupClient) statusClient() []world.Configuration {
	port := deployer.Port{
		Port:     8080,
		Name:     "http",
		Protocol: "TCP",
	}
	image := docker.Image{
		Repository: "bborbe/backup-status-client",
		Tag:        "2.0.0",
	}
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "status-client",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "client",
					Image: image,
					Requirement: &build.BackupStatusClient{
						Image: image,
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "50Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Args:  []k8s.Arg{"-logtostderr", "-v=1"},
					Ports: []deployer.Port{port},
					Env: []k8s.Env{
						{
							Name:  "PORT",
							Value: port.Port.String(),
						},
						{
							Name:  "SERVER",
							Value: "http://status-server:8080",
						},
					},
					LivenessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
					},
					ReadinessProbe: k8s.Probe{
						HttpGet: k8s.HttpGet{
							Path:   "/",
							Port:   port.Port,
							Scheme: "HTTP",
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "status-client",
			Ports:     []deployer.Port{port},
		},
		&deployer.IngressDeployer{
			Context:   b.Context,
			Namespace: "backup",
			Name:      "status-client",
			Port:      "http",
			Domains:   b.Domains,
		},
	}
}

func (b *BackupClient) Applier() (world.Applier, error) {
	return nil, nil
}
