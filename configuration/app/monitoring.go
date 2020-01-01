// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/container"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type MonitoringConfig struct {
	Name       k8s.MetadataName
	Subject    string
	GitRepoUrl container.GitRepoUrl
}

func (m *MonitoringConfig) Validate(ctx context.Context) error {
	if m.Subject == "" {
		return errors.New("Subject missing")
	}
	return validation.Validate(
		ctx,
		m.Name,
		m.GitRepoUrl,
	)
}

type Monitoring struct {
	Context         k8s.Context
	SmtpPassword    deployer.SecretValue
	GitSyncPassword deployer.SecretValue
	Configs         []MonitoringConfig
}

func (m *Monitoring) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		m.Context,
		m.SmtpPassword,
		m.GitSyncPassword,
	)
}

func (m *Monitoring) Children() []world.Configuration {
	configurations := []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: m.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "monitoring",
					Name:      "monitoring",
				},
			},
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   m.Context,
				Namespace: "monitoring",
				Name:      "monitoring",
				Secrets: deployer.Secrets{
					"git-sync-password": m.GitSyncPassword,
					"smtp-password":     m.SmtpPassword,
				},
			},
		),
	}
	for _, config := range m.Configs {
		configurations = append(configurations, &MonitoringDeployment{
			Context:    m.Context,
			Name:       config.Name,
			Subject:    config.Subject,
			GitRepoUrl: config.GitRepoUrl,
		})
	}
	return configurations
}

func (m *Monitoring) Applier() (world.Applier, error) {
	return nil, nil
}

type MonitoringDeployment struct {
	Context    k8s.Context
	Name       k8s.MetadataName
	Subject    string
	GitRepoUrl container.GitRepoUrl
}

func (m *MonitoringDeployment) Validate(ctx context.Context) error {
	if m.Subject == "" {
		return errors.New("Subject missing")
	}
	if m.GitRepoUrl == "" {
		return errors.New("GitRepoUrl missing")
	}
	return validation.Validate(
		ctx,
		m.Context,
		m.Name,
	)
}

func (m *MonitoringDeployment) Children() []world.Configuration {
	monitoringImage := docker.Image{
		Repository: "bborbe/monitoring",
		Tag:        "2.0.0",
	}
	mountName := k8s.MountName("data")
	return []world.Configuration{
		&deployer.DeploymentDeployer{
			Context:   m.Context,
			Namespace: "monitoring",
			Name:      m.Name,
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "monitoring",
					Image: monitoringImage,
					Requirement: &build.Monitoring{
						Image: monitoringImage,
					},
					Args: []k8s.Arg{
						"-logtostderr",
						"-v=1",
					},
					Env: []k8s.Env{
						{
							Name:  "CONFIG",
							Value: "/data/config.xml",
						},
						{
							Name:  "DRIVER",
							Value: "phantomjs",
						},
						{
							Name:  "DELAY",
							Value: "5m",
						},
						{
							Name:  "CONCURRENT",
							Value: "1",
						},
						{
							Name:  "ONE_TIME",
							Value: "false",
						},
						{
							Name:  "LOCK",
							Value: "/monitoring.lock",
						},
						{
							Name:  "SMTP_TLS",
							Value: "true",
						},
						{
							Name:  "SMTP_TLS_SKIP_VERIFY",
							Value: "true",
						},
						{
							Name:  "SMTP_USER",
							Value: "monitoring@benjamin-borbe.de",
						},
						{
							Name: "SMTP_PASSWORD",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "smtp-password",
									Name: "monitoring",
								},
							},
						},
						{
							Name:  "SMTP_HOST",
							Value: "mail.benjamin-borbe.de",
						},
						{
							Name:  "SMTP_PORT",
							Value: "465",
						},
						{
							Name:  "SENDER",
							Value: "monitoring@benjamin-borbe.de",
						},
						{
							Name:  "RECIPIENT",
							Value: "bborbe@rocketnews.de",
						},
						{
							Name:  "SUBJECT",
							Value: m.Subject,
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "100Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     mountName,
							Path:     "/data",
							ReadOnly: true,
						},
					},
				},
				&container.GitSync{
					MountName:                 mountName,
					GitRepoUrl:                m.GitRepoUrl,
					GitSyncUsername:           "bborbereadonly",
					GitSyncPasswordSecretName: "monitoring",
					GitSyncPasswordSecretPath: "git-sync-password",
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name:     mountName,
					EmptyDir: &k8s.PodVolumeEmptyDir{},
				},
			},
		},
	}
}

func (m *MonitoringDeployment) Applier() (world.Applier, error) {
	return nil, nil
}
