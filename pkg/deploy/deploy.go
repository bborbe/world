package deploy

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type Deployer struct {
	Context        world.Context
	Namespace      world.Namespace
	Domains        []world.Domain
	Args           []world.Arg
	Port           world.Port
	HostPort       world.HostPort
	Env            world.Env
	CpuLimit       world.CpuLimit
	MemoryLimit    world.MemoryLimit
	CpuRequest     world.CpuRequest
	MemoryRequest  world.MemoryRequest
	LivenessProbe  world.LivenessProbe
	ReadinessProbe world.ReadinessProbe
	Mounts         []world.Mount
	Image          world.Image
}

func (d *Deployer) BuildDeployment() k8s.Deployment {
	var mounts []k8s.VolumeMount
	var volumes []k8s.PodVolume
	for _, mount := range d.Mounts {
		volumes = append(volumes, k8s.PodVolume{
			Name: k8s.PodVolumeName(mount.Name),
			Nfs: k8s.PodNfs{
				Path:   k8s.PodNfsPath(mount.NfsPath),
				Server: k8s.PodNfsServer(mount.NfsServer),
			},
		})
		mounts = append(mounts, k8s.VolumeMount{
			MountPath: k8s.VolumeMountPath(mount.Target),
			Name:      k8s.VolumeName(mount.Name),
			ReadOnly:  k8s.VolumeReadOnly(mount.ReadOnly),
		})
	}
	container := k8s.PodContainer{
		Image: k8s.PodImage(d.Image.String()),
		Name:  k8s.PodName(d.Namespace.String()),
		Ports: []k8s.PodPort{
			{
				ContainerPort: k8s.PodPortContainerPort(d.Port),
				HostPort:      k8s.PodPortHostPort(d.HostPort),
				Name:          "http",
				Protocol:      "TCP",
			},
		},
		Resources: k8s.PodResources{
			Requests: k8s.ResourceList{
				"cpu":    d.CpuRequest.String(),
				"memory": d.MemoryRequest.String(),
			},
			Limits: k8s.ResourceList{
				"cpu":    d.CpuLimit.String(),
				"memory": d.MemoryLimit.String(),
			},
		},
		VolumeMounts: mounts,
	}
	for _, arg := range d.Args {
		container.Args = append(container.Args, k8s.PodArg(arg))
	}
	for k, v := range d.Env {
		container.Env = append(container.Env, k8s.PodEnv{
			Name:  k,
			Value: v,
		})
	}

	return k8s.Deployment{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Deployment",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(d.Namespace),
			Name:      k8s.Name(d.Namespace),
			Labels: k8s.Labels{
				"app": d.Namespace.String(),
			},
		},
		Spec: k8s.DeploymentSpec{
			Replicas:             1,
			RevisionHistoryLimit: 2,
			Selector: k8s.DeploymentSelector{
				MatchLabels: k8s.DeploymentMatchLabels{
					"app": d.Namespace.String(),
				},
			},
			Strategy: k8s.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: k8s.DeploymentStrategyRollingUpdate{
					MaxSurge:       1,
					MaxUnavailable: 1,
				},
			},
			Template: k8s.DeploymentTemplate{
				Metadata: k8s.Metadata{
					Labels: k8s.Labels{
						"app": d.Namespace.String(),
					},
				},
				Spec: k8s.PodSpec{
					Containers: []k8s.PodContainer{
						container,
					},
					Volumes: volumes,
				},
			},
		},
	}
}

func (i *Deployer) BuildIngress() k8s.Ingress {
	ingress := k8s.Ingress{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(i.Namespace),
			Name:      k8s.Name(i.Namespace),
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
			Annotations: k8s.Annotations{
				"kubernetes.io/ingress.class": "traefik",
				"traefik.frontend.priority":   "10000",
			},
		},
		Spec: k8s.IngressSpec{},
	}
	for _, domain := range i.Domains {
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.IngressRule{
			Host: k8s.IngressHost(domain),
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(i.Namespace),
							ServicePort: "web",
						},
					},
				},
			},
		})
	}
	return ingress
}

func (n *Deployer) BuildNamespace() k8s.Namespace {
	return k8s.Namespace{
		ApiVersion: "v1",
		Kind:       "Namespace",
		Metadata: k8s.Metadata{
			Name: k8s.Name(n.Namespace),
			Labels: k8s.Labels{
				"app": n.Namespace.String(),
			},
		},
	}
}

func (s *Deployer) BuildService() k8s.Service {
	return k8s.Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(s.Namespace),
			Name:      k8s.Name(s.Namespace),
			Labels: k8s.Labels{
				"app": s.Namespace.String(),
			},
		},
		Spec: k8s.ServiceSpec{
			Ports: []k8s.Port{
				{
					Name:       "web",
					Port:       k8s.PortNumber(s.Port),
					Protocol:   "TCP",
					TargetPort: "http",
				},
			},
			Selector: k8s.ServiceSelector{
				"app": s.Namespace.String(),
			},
		},
	}
}
