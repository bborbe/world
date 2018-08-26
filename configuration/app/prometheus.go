package app

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/configuration/cluster"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
)

type Prometheus struct {
	Cluster cluster.Cluster
}

func (t *Prometheus) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		t.Cluster,
	)
}

func (p *Prometheus) Children() []world.Configuration {
	nodeExporterImage := docker.Image{
		Repository: "bborbe/prometheus-node-exporter",
		Tag:        "v0.14.0",
	}
	return []world.Configuration{
		world.NewConfiguration().
			AddChildConfiguration(
				&build.PrometheusNodeExporter{
					Image: nodeExporterImage,
				}).
			WithApplier(
				&k8s.DaemonSetApplier{
					Context: p.Cluster.Context,
					DaemonSet: k8s.DaemonSet{
						ApiVersion: "apps/v1",
						Kind:       "DaemonSet",
						Metadata: k8s.Metadata{
							Name:      "prometheus-node-exporter",
							Namespace: "kube-system",
							Labels: k8s.Labels{
								"app":  "prometheus",
								"role": "node-exporter",
							},
						},
						Spec: k8s.DaemonSetSpec{
							Selector: k8s.Selector{
								MatchLabels: k8s.Labels{
									"app":  "prometheus",
									"role": "node-exporter",
								},
							},
							Template: k8s.PodTemplate{
								Metadata: k8s.Metadata{
									Annotations: k8s.Annotations{
										"prometheus.io/path":   "/metrics",
										"prometheus.io/port":   "9100",
										"prometheus.io/scheme": "http",
										"prometheus.io/scrape": "true",
									},
									Labels: k8s.Labels{
										"app":  "prometheus",
										"role": "node-exporter",
									},
								},
								Spec: k8s.PodSpec{
									HostPid: true,
									Containers: []k8s.Container{
										{
											Args: []k8s.Arg{
												"-collector.procfs=/host/proc",
												"-collector.sysfs=/host/sys",
												"-collector.filesystem.ignored-mount-points='^/(sys|proc|dev|host|etc)($|/)'",
											},
											Image: k8s.Image(nodeExporterImage.String()),
											Name:  "node-exporter",
											Ports: []k8s.ContainerPort{
												{
													Name:          "http",
													ContainerPort: 9100,
												},
											},
											Resources: k8s.Resources{
												Limits: k8s.ContainerResource{
													Cpu:    "100m",
													Memory: "50Mi",
												},
												Requests: k8s.ContainerResource{
													Cpu:    "10m",
													Memory: "10Mi",
												},
											},
											SecurityContext: k8s.SecurityContext{
												Privileged: true,
											},
										},
									},
								},
							},
						},
					},
				},
			),
	}
}

func (p *Prometheus) Applier() (world.Applier, error) {
	return nil, nil
}
