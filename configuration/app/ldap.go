package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Ldap struct {
	Context    k8s.Context
	NfsServer  deployer.MountNfsServer
	Tag        docker.Tag
	LdapSecret deployer.SecretValue
}

func (l *Ldap) Applier() world.Applier {
	return nil
}

func (l *Ldap) Children() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/openldap",
		Tag:        l.Tag,
	}
	ports := []deployer.Port{
		{
			Port:     389,
			Name:     "ldap",
			Protocol: "TCP",
		},
		{
			Port:     636,
			Name:     "ldaps",
			Protocol: "TCP",
		},
	}
	return []world.Configuration{
		&deployer.NamespaceDeployer{
			Context:   l.Context,
			Namespace: "ldap",
		},
		&deployer.SecretDeployer{
			Context:   l.Context,
			Namespace: "ldap",
			Secrets: deployer.Secrets{
				"secret": l.LdapSecret,
			},
		},
		&deployer.DeploymentDeployer{
			Context: l.Context,
			Requirements: []world.Configuration{
				&build.Openldap{
					Image: image,
				},
			},
			Namespace: "ldap",
			Containers: []deployer.DeploymentDeployerContainer{
				{
					Name: "ldap",
					Env: []k8s.Env{
						{
							Name:  "LDAP_SUFFIX",
							Value: "dc=benjamin-borbe,dc=de",
						},
						{
							Name:  "LDAP_ROOTDN",
							Value: "cn=root,dc=benjamin-borbe,dc=de",
						},
						{
							Name: "LDAP_SECRET",
							ValueFrom: k8s.ValueFrom{
								SecretKeyRef: k8s.SecretKeyRef{
									Key:  "secret",
									Name: "ldap",
								},
							},
						},
					},
					Image:         image,
					Ports:         ports,
					CpuLimit:      "100m",
					MemoryLimit:   "50Mi",
					CpuRequest:    "10m",
					MemoryRequest: "10Mi",
					Mounts: []deployer.Mount{
						{
							Name:   "ldap",
							Target: "/var/lib/openldap/openldap-data",
						},
					},
				},
			},
			Volumes: []deployer.Volume{
				{
					Name:      "ldap",
					NfsPath:   "/data/ldap",
					NfsServer: l.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   l.Context,
			Namespace: "ldap",
			Name:      "ldap",
			Ports:     ports,
		},
	}
}

func (l *Ldap) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate ldap app ...")
	if l.Context == "" {
		return errors.New("Context missing")
	}
	if l.NfsServer == "" {
		return errors.New("NfsServer missing")
	}
	if l.Tag == "" {
		return errors.New("Tag missing")
	}
	if l.LdapSecret == nil {
		return errors.New("LdapSecret missing")
	}
	if err := l.LdapSecret.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate LdapSecret failed")
	}
	return nil
}
