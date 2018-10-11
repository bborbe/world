package app

import (
	"context"
	"strconv"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/deployer"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type NfsProvisioner struct {
	Context             k8s.Context
	HostPath            k8s.PodHostPath
	DefaultStorageClass bool
}

func (n *NfsProvisioner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Context,
		n.HostPath,
	)
}

func (n *NfsProvisioner) Applier() (world.Applier, error) {
	return nil, nil
}

func (n *NfsProvisioner) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/nfs-provisioner",
		Tag:        "v1.0.9",
	}

	ports := []deployer.Port{
		{
			Port:     2049,
			Protocol: "TCP",
			Name:     "nfs",
		},
		{
			Port:     20048,
			Protocol: "TCP",
			Name:     "mountd",
		},
		{
			Port:     111,
			Protocol: "TCP",
			Name:     "rpcbind",
		},
		{
			Port:     111,
			Protocol: "UDP",
			Name:     "rpcbind-udp",
		},
	}
	return []world.Configuration{
		&k8s.ServiceaccountConfiguration{
			Context: n.Context,
			Serviceaccount: k8s.Serviceaccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "nfs-provisioner",
				},
			},
		},
		&k8s.StorageClassConfiguration{
			Context: n.Context,
			StorageClass: k8s.StorageClass{
				ApiVersion: "storage.k8s.io/v1",
				Kind:       "StorageClass",
				Metadata: k8s.Metadata{
					Namespace: "kube-system",
					Name:      "nfs",
					Annotations: map[string]string{
						"storageclass.kubernetes.io/is-default-class": strconv.FormatBool(n.DefaultStorageClass),
					},
				},
				Provisioner: "example.com/nfs",
				Parameters: map[string]string{
					"mountOptions": "vers=4.1",
				},
			},
		},
		&deployer.DeploymentDeployer{
			Context:   n.Context,
			Namespace: "kube-system",
			Name:      "nfs-provisioner",
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Containers: []deployer.HasContainer{
				&deployer.DeploymentDeployerContainer{
					Name:  "nfs-provisioner",
					Image: image,
					Requirement: &build.NfsProvisioner{
						Image: image,
					},
					Ports: ports,
					SecurityContext: k8s.SecurityContext{
						Capabilities: map[string][]string{
							"add": {
								"DAC_READ_SEARCH",
								"SYS_RESOURCE",
							},
						},
					},
					Args: []k8s.Arg{"-provisioner=example.com/nfs"},
					Env: []k8s.Env{
						{
							Name: "POD_IP",
							ValueFrom: k8s.ValueFrom{
								FieldRef: k8s.FieldRef{
									FieldPath: "status.podIP",
								},
							},
						},
						{
							Name:  "SERVICE_NAME",
							Value: "nfs-provisioner",
						},
						{
							Name: "POD_NAMESPACE",
							ValueFrom: k8s.ValueFrom{
								FieldRef: k8s.FieldRef{
									FieldPath: "metadata.namespace",
								},
							},
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name: "export-volume",
							Path: "/export",
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "2000m",
							Memory: "2000Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "export-volume",
					Host: k8s.PodVolumeHost{
						Path: n.HostPath,
					},
				},
			},
		},
		&deployer.ServiceDeployer{
			Context:   n.Context,
			Namespace: "kube-system",
			Name:      "nfs-provisioner",
			Ports:     ports,
		},
	}
}
