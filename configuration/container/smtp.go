// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package container

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type SmtpHostname string

func (s SmtpHostname) String() string {
	return string(s)
}

func (s SmtpHostname) Validate(ctx context.Context) error {
	if s == "" {
		return errors.New("NamespaceName missing")
	}
	return nil
}

type Smtp struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Hostname     SmtpHostname
	SmtpPassword deployer.SecretValue
	SmtpUsername deployer.SecretValue
}

func (s *Smtp) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.Namespace,
		s.Hostname,
		s.SmtpUsername,
		s.SmtpPassword,
	)
}

func (s *Smtp) Requirements() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   s.Context,
				Namespace: s.Namespace,
				Name:      "smtp",
				Secrets: deployer.Secrets{
					"username": s.SmtpUsername,
					"password": s.SmtpPassword,
				},
			},
		),
	}
}

func (s *Smtp) Container() *deployer.DeploymentDeployerContainer {
	image := docker.Image{
		Repository: "bborbe/smtp",
		Tag:        "1.2.1",
	}
	return &deployer.DeploymentDeployerContainer{
		Name:  "smtp",
		Image: image,
		Requirement: &build.Smtp{
			Image: image,
		},
		Ports: []deployer.Port{
			{
				Port:     25,
				Protocol: "TCP",
				Name:     "smtp",
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
		Env: []k8s.Env{
			{
				Name:  "HOSTNAME",
				Value: s.Hostname.String(),
			},
			{
				Name:  "RELAY_SMTP_PORT",
				Value: "25",
			},
			{
				Name:  "RELAY_SMTP_SERVER",
				Value: "mail.benjamin-borbe.de",
			},
			{
				Name:  "RELAY_SMTP_TLS",
				Value: "false",
			},
			{
				Name:  "ALLOWED_SENDER_DOMAINS",
				Value: "",
			},
			{
				Name:  "ALLOWED_NETWORKS",
				Value: "",
			},
			{
				Name: "RELAY_SMTP_USERNAME",
				ValueFrom: k8s.ValueFrom{
					SecretKeyRef: k8s.SecretKeyRef{
						Key:  "username",
						Name: "smtp",
					},
				},
			},
			{
				Name: "RELAY_SMTP_PASSWORD",
				ValueFrom: k8s.ValueFrom{
					SecretKeyRef: k8s.SecretKeyRef{
						Key:  "password",
						Name: "smtp",
					},
				},
			},
		},
	}
}
