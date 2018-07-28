package builder

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type DeploymentBuilder struct {
	Namespace     world.Namespace
	Args          []world.Arg
	Port          world.Port
	HostPort      world.HostPort
	Env           world.Env
	CpuLimit      world.CpuLimit
	MemoryLimit   world.MemoryLimit
	CpuRequest    world.CpuRequest
	MemoryRequest world.MemoryRequest
	Mounts        []world.Mount
	Image         world.Image
}

func (d *DeploymentBuilder) Build() k8s.Deployment {
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
