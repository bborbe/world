package app

import (
	"context"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Ldap struct {
	Cluster      cluster.Cluster
	Tag          docker.Tag
	LdapPassword deployer.SecretValue
}

func (d *Ldap) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Tag,
		d.LdapPassword,
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
	ldapPort := deployer.Port{
		Port:     389,
		Name:     "ldap",
		Protocol: "TCP",
	}
	ldapsPort := deployer.Port{
		Port:     636,
		Name:     "ldaps",
		Protocol: "TCP",
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
				"secret": l.LdapPassword,
			},
		},
		&deployer.DeploymentDeployer{
			Context:   l.Cluster.Context,
			Namespace: "ldap",
			Name:      "ldap",
			Strategy: k8s.DeploymentStrategy{
				Type: "Recreate",
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
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
					Ports: []deployer.Port{ldapPort, ldapsPort},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "500m",
							Memory: "75Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "100m",
							Memory: "25Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "ldap",
							Path: "/var/lib/openldap/openldap-data",
						},
					},
					LivenessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: ldapPort.Port,
						},
						InitialDelaySeconds: 60,
						SuccessThreshold:    1,
						FailureThreshold:    5,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: k8s.Probe{
						TcpSocket: k8s.TcpSocket{
							Port: ldapPort.Port,
						},
						InitialDelaySeconds: 3,
						TimeoutSeconds:      5,
						PeriodSeconds:       10,
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
			Ports:     []deployer.Port{ldapPort, ldapsPort},
		},
	}
}
