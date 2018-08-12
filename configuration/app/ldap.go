package app

import (
	"context"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Ldap struct {
	Context            k8s.Context
	NfsServer          world.MountNfsServer
	Tag                docker.Tag
	TeamvaultConnector connector.Connector
}

func (d *Ldap) Applier() world.Applier {
	return nil
}

func (d *Ldap) Childs() []world.Configuration {
	image := docker.Image{
		Registry:   "docker.io",
		Repository: "bborbe/openldap",
		Tag:        d.Tag,
	}
	ports := []world.Port{
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
			Context:   d.Context,
			Namespace: "ldap",
		},
		&deployer.SecretDeployer{
			Context:   d.Context,
			Namespace: "ldap",
			Secrets: world.Secrets{
				"secret": &world.SecretFromTeamvault{
					TeamvaultConnector: d.TeamvaultConnector,
					TeamvaultKey:       "MOPMLG",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context: d.Context,
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
					Mounts: []world.Mount{
						{
							Name:   "ldap",
							Target: "/var/lib/openldap/openldap-data",
						},
					},
				},
			},
			Volumes: []world.Volume{
				{
					Name:      "ldap",
					NfsPath:   "/data/ldap",
					NfsServer: d.NfsServer,
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   d.Context,
			Namespace: "ldap",
			Name:      "ldap",
			Ports:     ports,
		},
	}
}

func (d *Ldap) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate ldap app ...")
	if d.Context == "" {
		return errors.New("context missing in ldap app")
	}
	if d.NfsServer == "" {
		return errors.New("nfs-server missing in ldap app")
	}
	if d.TeamvaultConnector == nil {
		return errors.New("teamvault-connector missing in ldap app")
	}
	return nil
}
