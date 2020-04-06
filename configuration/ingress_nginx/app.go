// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ingress_nginx

import (
	"context"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type App struct {
	Context k8s.Context
}

func (a *App) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		a.Context,
	)
}
func (a *App) Children() []world.Configuration {
	image := docker.Image{
		Repository: "bborbe/nginx-ingress-controller",
		Tag:        "0.30.0", // https://quay.io/repository/kubernetes-ingress-controller/nginx-ingress-controller?tag=latest&tab=tags
	}
	return []world.Configuration{
		&k8s.NamespaceConfiguration{
			Context: a.Context,
			Namespace: k8s.Namespace{
				ApiVersion: "v1",
				Kind:       "Namespace",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "ingress-nginx",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: a.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "nginx-configuration",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
			},
		},
		&k8s.ServiceaccountConfiguration{
			Context: a.Context,
			Serviceaccount: k8s.ServiceAccount{
				ApiVersion: "v1",
				Kind:       "ServiceAccount",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "nginx-ingress-serviceaccount",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
			},
		},
		&k8s.ClusterRoleConfiguration{
			Context: a.Context,
			ClusterRole: k8s.ClusterRole{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
				Metadata: k8s.Metadata{
					Namespace: "",
					Name:      "nginx-ingress-clusterrole",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
					Annotations: k8s.Annotations(nil),
				},
				Rules: []k8s.PolicyRule{
					{
						ApiGroups: []string{
							"",
						},
						NonResourceURLs: []string(nil),
						ResourceNames:   []string(nil),
						Resources: []string{
							"configmaps",
							"endpoints",
							"nodes",
							"pods",
							"secrets",
						},
						Verbs: []string{
							"list",
							"watch",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						NonResourceURLs: []string(nil),
						ResourceNames:   []string(nil),
						Resources: []string{
							"nodes",
						},
						Verbs: []string{
							"get",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"services",
						},
						Verbs: []string{
							"get",
							"list",
							"watch",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"events",
						},
						Verbs: []string{
							"create",
							"patch",
						},
					},
					{
						ApiGroups: []string{
							"extensions",
							"networking.k8s.io",
						},
						Resources: []string{
							"ingresses",
						},
						Verbs: []string{
							"get",
							"list",
							"watch",
						},
					},
					{
						ApiGroups: []string{
							"extensions",
							"networking.k8s.io",
						},
						Resources: []string{
							"ingresses/status",
						},
						Verbs: []string{
							"update",
						},
					},
				},
			},
		},
		&k8s.ClusterRoleBindingConfiguration{
			Context: a.Context,
			ClusterRoleBinding: k8s.ClusterRoleBinding{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
				Metadata: k8s.Metadata{
					Namespace: "",
					Name:      "nginx-ingress-clusterrole-nisa-binding",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
				Subjects: []k8s.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "nginx-ingress-serviceaccount",
						Namespace: "ingress-nginx",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "ClusterRole",
					Name:     "nginx-ingress-clusterrole",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		&k8s.RoleConfiguration{
			Context: a.Context,
			Role: k8s.Role{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "Role",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "nginx-ingress-role",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
				Rules: []k8s.PolicyRule{
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"configmaps",
							"pods",
							"secrets",
							"namespaces",
						},
						Verbs: []string{
							"get",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						ResourceNames: []string{
							"ingress-controller-leader-nginx",
						},
						Resources: []string{
							"configmaps",
						},
						Verbs: []string{
							"get",
							"update",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"configmaps",
						},
						Verbs: []string{
							"create",
						},
					},
					{
						ApiGroups: []string{
							"",
						},
						Resources: []string{
							"endpoints",
						},
						Verbs: []string{
							"get",
						},
					},
				},
			},
		},
		&k8s.RoleBindingConfiguration{
			Context: a.Context,
			RoleBinding: k8s.RoleBinding{
				ApiVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "RoleBinding",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "nginx-ingress-role-nisa-binding",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
				Subjects: []k8s.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "nginx-ingress-serviceaccount",
						Namespace: "ingress-nginx",
					},
				},
				RoleRef: k8s.RoleRef{
					Kind:     "Role",
					Name:     "nginx-ingress-role",
					ApiGroup: "rbac.authorization.k8s.io",
				},
			},
		},
		&Build{
			Image: image,
		},
		&k8s.DeploymentConfiguration{
			Context: a.Context,
			Deployment: k8s.Deployment{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "nginx-ingress-controller",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx",
					},
				},
				Spec: k8s.DeploymentSpec{
					Replicas: 1,
					Selector: k8s.LabelSelector{
						MatchLabels: k8s.Labels{
							"app.kubernetes.io/name":    "ingress-nginx",
							"app.kubernetes.io/part-of": "ingress-nginx",
						},
					},
					Template: k8s.PodTemplate{
						Metadata: k8s.Metadata{
							Labels: k8s.Labels{
								"app.kubernetes.io/name":    "ingress-nginx",
								"app.kubernetes.io/part-of": "ingress-nginx",
							},
							Annotations: k8s.Annotations{
								"prometheus.io/port":   "10254",
								"prometheus.io/scrape": "true",
							},
						},
						Spec: k8s.PodSpec{
							Containers: []k8s.Container{
								{
									Name:  "nginx-ingress-controller",
									Image: k8s.Image(image.String()),
									Args: []k8s.Arg{
										"/nginx-ingress-controller",
										"--configmap=$(POD_NAMESPACE)/nginx-configuration",
										"--tcp-services-configmap=$(POD_NAMESPACE)/tcp-services",
										"--udp-services-configmap=$(POD_NAMESPACE)/udp-services",
										"--publish-service=$(POD_NAMESPACE)/ingress-nginx",
										"--annotations-prefix=nginx.ingress.kubernetes.io",
									},
									Env: []k8s.Env{
										{
											Name: "POD_NAME",
											ValueFrom: k8s.ValueFrom{
												FieldRef: k8s.FieldRef{
													FieldPath: "metadata.name",
												},
											},
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
									Ports: k8s.ContainerPorts{
										k8s.ContainerPort{
											ContainerPort: 80,
											HostPort:      80,
											Name:          "http",
										},
										k8s.ContainerPort{
											ContainerPort: 443,
											HostPort:      443,
											Name:          "https",
										},
									},
									Resources: k8s.Resources{
										Limits: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "90Mi",
										},
										Requests: k8s.ContainerResource{
											Cpu:    "100m",
											Memory: "90Mi",
										},
									},
									ReadinessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/healthz",
											Port:   10254,
											Scheme: "HTTP",
										},
										SuccessThreshold: 1,
										FailureThreshold: 3,
										TimeoutSeconds:   10,
										PeriodSeconds:    10,
									},
									LivenessProbe: k8s.Probe{
										HttpGet: k8s.HttpGet{
											Path:   "/healthz",
											Port:   10254,
											Scheme: "HTTP",
										},
										InitialDelaySeconds: 10,
										SuccessThreshold:    1,
										FailureThreshold:    3,
										TimeoutSeconds:      10,
										PeriodSeconds:       10,
									},
									SecurityContext: k8s.SecurityContext{
										AllowPrivilegeEscalation: true,
										RunAsUser:                101,
										Capabilities: k8s.SecurityContextCapabilities{
											"add": []string{
												"NET_BIND_SERVICE",
											},
											"drop": []string{
												"ALL",
											},
										},
									},
								},
							},
							ServiceAccountName:            "nginx-ingress-serviceaccount",
							TerminationGracePeriodSeconds: 300,
						},
					},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: a.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{Namespace: "ingress-nginx",
					Name: "tcp-services",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx"},
				},
			},
		},
		&k8s.ConfigMapConfiguration{
			Context: a.Context,
			ConfigMap: k8s.ConfigMap{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				Metadata: k8s.Metadata{
					Namespace: "ingress-nginx",
					Name:      "udp-services",
					Labels: k8s.Labels{
						"app.kubernetes.io/name":    "ingress-nginx",
						"app.kubernetes.io/part-of": "ingress-nginx"},
				},
			},
		},
	}
}
func (a *App) Applier() (world.Applier, error) {
	return nil, nil
}
