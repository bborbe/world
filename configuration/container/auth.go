// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package container

import (
	"context"
	"fmt"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Auth struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Port         k8s.PortNumber
	TargetPort   k8s.PortNumber
	Secret       deployer.SecretValue
	LdapUsername deployer.SecretValue
	LdapPassword deployer.SecretValue
}

func (a *Auth) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		a.Context,
		a.Namespace,
		a.Port,
		a.TargetPort,
		a.Secret,
		a.LdapUsername,
		a.LdapPassword,
	)
}

func (a *Auth) Requirements() []world.Configuration {
	return []world.Configuration{
		&build.AuthHttpProxy{
			Image: a.image(),
		},
		world.NewConfiguraionBuilder().WithApplier(
			&deployer.SecretApplier{
				Context:   a.Context,
				Namespace: a.Namespace,
				Name:      "auth",
				Secrets: deployer.Secrets{
					"ldap-user":     a.LdapUsername,
					"ldap-password": a.LdapPassword,
					"auth-secret":   a.Secret,
				},
			},
		),
	}
}

func (a *Auth) image() docker.Image {
	return docker.Image{
		Repository: "bborbe/auth-http-proxy",
		Tag:        "3.2.1",
	}
}

func (a *Auth) Container() k8s.Container {
	return k8s.Container{
		Name:  "auth",
		Image: k8s.Image(a.image().String()),
		Args: []k8s.Arg{
			"-logtostderr",
			"-v=2",
		},
		Ports: []k8s.ContainerPort{
			{
				ContainerPort: a.Port,
				Protocol:      "TCP",
				Name:          "http-auth",
			},
		},
		Env: []k8s.Env{
			{
				Name:  "PORT",
				Value: a.Port.String(),
			},
			{
				Name:  "DEBUG",
				Value: "false",
			},
			{
				Name:  "KIND",
				Value: "html",
			},
			{
				Name:  "TARGET_ADDRESS",
				Value: fmt.Sprintf("127.0.0.1:%d", a.TargetPort),
			},
			{
				Name: "SECRET",
				ValueFrom: k8s.ValueFrom{
					SecretKeyRef: k8s.SecretKeyRef{
						Name: "auth",
						Key:  "auth-secret",
					},
				},
			},
			{
				Name:  "VERIFIER",
				Value: "ldap",
			},
			{
				Name:  "LDAP_HOST",
				Value: "ldap.ldap.svc.cluster.local",
			},
			{
				Name:  "LDAP_SERVERNAME",
				Value: "ldap.ldap.svc.cluster.local",
			},
			{
				Name:  "LDAP_SKIP_TLS",
				Value: "true",
			},
			{
				Name:  "LDAP_PORT",
				Value: "389",
			},
			{
				Name: "LDAP_BIND_DN",
				ValueFrom: k8s.ValueFrom{
					SecretKeyRef: k8s.SecretKeyRef{
						Name: "auth",
						Key:  "ldap-user",
					},
				},
			},
			{
				Name: "LDAP_BIND_PASSWORD",
				ValueFrom: k8s.ValueFrom{
					SecretKeyRef: k8s.SecretKeyRef{
						Name: "auth",
						Key:  "ldap-password",
					},
				},
			},
			{
				Name:  "LDAP_BASE_DN",
				Value: "dc=benjamin-borbe,dc=de",
			},
			{
				Name:  "LDAP_USER_DN",
				Value: "ou=users",
			},
			{
				Name:  "LDAP_USER_FILTER",
				Value: "(uid=%s)",
			},
			{
				Name:  "LDAP_GROUP_DN",
				Value: "ou=groups",
			},
			{
				Name:  "LDAP_GROUP_FILTER",
				Value: "(member=uid=%s,ou=users,dc=benjamin-borbe,dc=de)",
			},
			{
				Name:  "REQUIRED_GROUPS",
				Value: "Admins",
			},
		},
		Resources: k8s.Resources{
			Limits: k8s.ContainerResource{
				Cpu:    "500m",
				Memory: "100Mi",
			},
			Requests: k8s.ContainerResource{
				Cpu:    "100m",
				Memory: "50Mi",
			},
		},
		ReadinessProbe: k8s.Probe{
			HttpGet: k8s.HttpGet{
				Path:   "/readiness",
				Port:   a.Port,
				Scheme: "HTTP",
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      5,
		},
		LivenessProbe: k8s.Probe{
			HttpGet: k8s.HttpGet{
				Path:   "/healthz",
				Port:   a.Port,
				Scheme: "HTTP",
			},
			InitialDelaySeconds: 30,
			SuccessThreshold:    1,
			FailureThreshold:    5,
			TimeoutSeconds:      5,
		},
	}
}
