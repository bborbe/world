package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Ldap struct {
	Cluster    cluster.Cluster
	Tag        docker.Tag
	LdapSecret deployer.SecretValue
}

func (d *Ldap) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Tag,
		d.LdapSecret,
	)
}

func (l *Ldap) Applier() (world.Applier, error) {
	return nil, nil
}

func (l *Ldap) Children() []world.Configuration {
	image := docker.Image{
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
			Context:   l.Cluster.Context,
			Namespace: "ldap",
		},
		&deployer.SecretDeployer{
			Context:   l.Cluster.Context,
			Namespace: "ldap",
			Name:      "ldap",
			Secrets: deployer.Secrets{
				"secret": l.LdapSecret,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   l.Cluster.Context,
			Namespace: "ldap",
			Name:      "ldap",
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
					Image: image,
					Requirement: &build.Openldap{
						Image: image,
					},
					Ports: ports,
					Resources: k8s.PodResources{
						Limits: k8s.Resources{
							Cpu:    "500m",
							Memory: "75Mi",
						},
						Requests: k8s.Resources{
							Cpu:    "100m",
							Memory: "25Mi",
						},
					},
					Mounts: []k8s.VolumeMount{
						{
							Name: "ldap",
							Path: "/var/lib/openldap/openldap-data",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "ldap",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/ldap",
						Server: l.Cluster.NfsServer,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   l.Cluster.Context,
			Namespace: "ldap",
			Name:      "ldap",
			Ports:     ports,
		},
	}
}
