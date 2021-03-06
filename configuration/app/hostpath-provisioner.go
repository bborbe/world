// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"strconv"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type HostPathProvisioner struct {
	Context             k8s.Context
	HostPath            k8s.PodHostPath
	DefaultStorageClass bool
	DisableRBAC         bool
}

func (h *HostPathProvisioner) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		h.Context,
		h.HostPath,
	)
}

func (h *HostPathProvisioner) Applier() (world.Applier, error) {
	return nil, nil
}

func (h *HostPathProvisioner) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/hostpath-provisioner",
		Tag:        "1.1.0",
	}
	namespace := "hostpath"
	result := []world.Configuration{
		&build.HostPathProvisioner{
			Image: image,
		},
		&k8s.NamespaceConfiguration{
			Context: h.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: k8s.NamespaceName(namespace),
					Name:      k8s.MetadataName(namespace),
				},
			},
		},
		&k8s.ServiceaccountConfiguration{
			Context: h.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: k8s.NamespaceName(namespace),
					Name:      "hostpath-provisioner",
				},
			},
		},
		&k8s.StorageClassConfiguration{
			Context: h.Context,
			StorageClass: k8s.StorageClass{
				ApiVersion: "storage.k8s.io/v1",
				Kind:       "StorageClass",
				Metadata: k8s.Metadata{
					Namespace: k8s.NamespaceName(namespace),
					Name:      "hostpath",
					Annotations: map[string]string{
						"storageclass.kubernetes.io/is-default-class": strconv.FormatBool(h.DefaultStorageClass),
					},
				},
				Provisioner: "hostpath",
			},
		},
		&k8s.DeploymentConfiguration{
			Context: h.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: k8s.NamespaceName(namespace),
					Name:      "hostpath-provisioner",
				},
				Spec: k8s.DeploymentSpec{
					Replicas: 1,
					Strategy: k8s.DeploymentStrategy{
						Type: "RollingUpdate",
						RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
							MaxUnavailable: 1,
						},
					},
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app": "hostpath-provisioner",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app": "hostpath-provisioner",
							},
						},
						Spec: k8s.PodSpec{
							ServiceAccountName: "hostpath-provisioner",
							Containers: []k8s.Container{
								{
									Name:            "hostpath-provisioner",
									Image:           k8s.Image(image.String()),
									ImagePullPolicy: "IfNotPresent",
									SecurityContext: k8s.SecurityContext{
										Capabilities: map[string][]string{
											"add": {
												"DAC_READ_SEARCH",
												"SYS_RESOURCE",
											},
										},
									},
									Env: []k8s.Env{
										{
											Name: "NODE_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "spec.nodeName",
												},
											},
										},
										{
											Name:  "PV_DIR",
											Value: h.HostPath.String(),
										},
										//{
										//	Name:  "PV_RECLAIM_POLICY",
										//	Value: "Retain",
										//},
									},
									VolumeMounts: []k8s.ContainerMount{
										{
											Name: "data",
											Path: k8s.ContainerMountPath(h.HostPath),
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "200m",
											Memory: "200Mi",
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
									Name: "data",
									Host: k8s.PodVolumeHost{
										Path: h.HostPath,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if !h.DisableRBAC {
		result = append(result,
			&k8s.ClusterRoleConfiguration{
				Context: h.Context,
				ClusterRole: k8s.ClusterRole{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRole",
					Metadata: k8s.Metadata{
						Name: "hostpath-provisioner",
					},
					Rules: []k8s.PolicyRule{
						{
							ApiGroups: []string{"*"},
							Resources: []string{"*"},
							Verbs:     []string{"*"},
						},
						{
							NonResourceURLs: []string{"*"},
							Verbs:           []string{"*"},
						},
					},
				},
			},
			&k8s.ClusterRoleBindingConfiguration{
				Context: h.Context,
				ClusterRoleBinding: k8s.ClusterRoleBinding{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Kind:       "ClusterRoleBinding",
					Metadata: k8s.Metadata{
						Name: "hostpath-provisioner",
					},
					Subjects: []k8s.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      "hostpath-provisioner",
							Namespace: namespace,
						},
					},
					RoleRef: k8s.RoleRef{
						ApiGroup: "rbac.authorization.k8s.io",
						Kind:     "ClusterRole",
						Name:     "hostpath-provisioner",
					},
				},
			},
		)
	}

	return result
}
