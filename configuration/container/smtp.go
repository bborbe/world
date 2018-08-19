package container

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type SmtpProvider struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Hostname     string
	SmtpPassword deployer.SecretValue
	SmtpUsername deployer.SecretValue
}

func (s *SmtpProvider) Container() deployer.DeploymentDeployerContainer {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/smtp",
		Tag:        "1.2.1",
	}
	return deployer.DeploymentDeployerContainer{
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
		CpuLimit:      "250m",
		MemoryLimit:   "100Mi",
		CpuRequest:    "10m",
		MemoryRequest: "10Mi",
		Env: []k8s.Env{
			{
				Name:  "HOSTNAME",
				Value: s.Hostname,
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

func (s *SmtpProvider) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("Context missing")
	}
	if s.Namespace == "" {
		return errors.New("Namespace context missing")
	}
	if s.Hostname == "" {
		return errors.New("Hostname missing")
	}
	if err := s.SmtpUsername.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	if err := s.SmtpPassword.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate failed")
	}
	return nil
}

func (s *SmtpProvider) Requirements() []world.Configuration {
	return []world.Configuration{
		&deployer.SecretDeployer{
			Context:   s.Context,
			Namespace: s.Namespace,
			Name:      "smtp",
			Secrets: deployer.Secrets{
				"username": s.SmtpUsername,
				"password": s.SmtpPassword,
			},
		},
	}
}
