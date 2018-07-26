package deploy

import (
	"context"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context        world.ClusterContext
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

func (d *Deployer) Deploy(ctx context.Context) error {
	glog.V(2).Infof("deploy %s to %s ...", d.Namespace, d.Context)

	namespace := &k8s.Deployer{
		Context: d.Context,
		Data:    d.namespace(),
	}
	if err := namespace.Deploy(ctx); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	service := &k8s.Deployer{
		Context: d.Context,
		Data:    d.service(),
	}
	if err := service.Deploy(ctx); err != nil {
		return errors.Wrap(err, "apply service failed")
	}

	if len(d.Domains) > 0 {
		ingress := &k8s.Deployer{
			Context: d.Context,
			Data:    d.ingress(),
		}
		if err := ingress.Deploy(ctx); err != nil {
			return errors.Wrap(err, "apply ingress failed")
		}
	}

	deployment := &k8s.Deployer{
		Context: d.Context,
		Data:    d.deployment(),
	}
	if err := deployment.Deploy(ctx); err != nil {
		return errors.Wrap(err, "apply deployment failed")
	}

	glog.V(2).Infof("deploy %s to %s finished", d.Namespace, d.Context)
	return nil
}

func (d *Deployer) Validate(ctx context.Context) error {
	if d.Context == "" {
		return errors.New("context missing")
	}
	if d.Namespace == "" {
		return errors.New("namespace missing")
	}
	for _, domain := range d.Domains {
		if domain == "" {
			return errors.New("domain empty")
		}
	}
	if d.Port <= 0 || d.Port > 65535 {
		return errors.New("port missing")
	}
	if d.CpuLimit == "" {
		return errors.New("cpu limit missing")
	}
	if d.MemoryLimit == "" {
		return errors.New("memory limit missing")
	}
	if d.CpuRequest == "" {
		return errors.New("cpu request missing")
	}
	if d.MemoryRequest == "" {
		return errors.New("memory request missing")
	}
	for _, mount := range d.Mounts {
		if err := mount.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deployer) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (d *Deployer) ingress() k8s.Ingress {
	ingress := k8s.Ingress{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Ingress",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(d.Namespace),
			Name:      k8s.Name(d.Namespace),
			Labels: k8s.Labels{
				"app": d.Namespace.String(),
			},
			Annotations: k8s.Annotations{
				"kubernetes.io/ingress.class": "traefik",
				"traefik.frontend.priority":   "10000",
			},
		},
		Spec: k8s.IngressSpec{},
	}
	for _, domain := range d.Domains {
		ingress.Spec.Rules = append(ingress.Spec.Rules, k8s.IngressRule{
			Host: k8s.IngressHost(domain),
			Http: k8s.IngressHttp{
				Paths: []k8s.IngressPath{
					{
						Path: "/",
						Backends: k8s.IngressBackend{
							ServiceName: k8s.IngressBackendServiceName(d.Namespace),
							ServicePort: "web",
						},
					},
				},
			},
		})
	}
	return ingress
}

func (n *Deployer) namespace() k8s.Namespace {
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

func (n *Deployer) service() k8s.Service {
	return k8s.Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(n.Namespace),
			Name:      k8s.Name(n.Namespace),
			Labels: k8s.Labels{
				"app": n.Namespace.String(),
			},
		},
		Spec: k8s.ServiceSpec{
			Ports: []k8s.Port{
				{
					Name:       "web",
					Port:       k8s.PortNumber(n.Port),
					Protocol:   "TCP",
					TargetPort: "http",
				},
			},
			Selector: k8s.ServiceSelector{
				"app": n.Namespace.String(),
			},
		},
	}
}

func (n *Deployer) deployment() k8s.Deployment {
	var mounts []k8s.VolumeMount
	var volumes []k8s.PodVolume
	for _, mount := range n.Mounts {
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
		Image: k8s.PodImage(n.Image.String()),
		Name:  k8s.PodName(n.Namespace.String()),
		Ports: []k8s.PodPort{
			{
				ContainerPort: k8s.PodPortContainerPort(n.Port),
				HostPort:      k8s.PodPortHostPort(n.HostPort),
				Name:          "http",
				Protocol:      "TCP",
			},
		},
		Resources: k8s.PodResources{
			Requests: k8s.ResourceList{
				"cpu":    n.CpuRequest.String(),
				"memory": n.MemoryRequest.String(),
			},
			Limits: k8s.ResourceList{
				"cpu":    n.CpuLimit.String(),
				"memory": n.MemoryLimit.String(),
			},
		},
		VolumeMounts: mounts,
	}
	for _, arg := range n.Args {
		container.Args = append(container.Args, k8s.PodArg(arg))
	}
	for k, v := range n.Env {
		container.Env = append(container.Env, k8s.PodEnv{
			Name:  k,
			Value: v,
		})
	}

	deployment := k8s.Deployment{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Deployment",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(n.Namespace),
			Name:      k8s.Name(n.Namespace),
			Labels: k8s.Labels{
				"app": n.Namespace.String(),
			},
		},
		Spec: k8s.DeploymentSpec{
			Replicas:             1,
			RevisionHistoryLimit: 2,
			Selector: k8s.DeploymentSelector{
				MatchLabels: k8s.DeploymentMatchLabels{
					"app": n.Namespace.String(),
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
						"app": n.Namespace.String(),
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
	return deployment
}
