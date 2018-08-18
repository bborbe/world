package deployer

import (
	"context"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/model"
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Port struct {
	Name     string
	Port     int
	HostPort int
	Protocol string
}

type SecretValue interface {
	Value() ([]byte, error)
	Validate(ctx context.Context) error
}

type SecretFromTeamvault struct {
	TeamvaultConnector connector.Connector
	TeamvaultKey       model.TeamvaultKey
}

func (s *SecretFromTeamvault) Value() ([]byte, error) {
	teamvaultPassword, err := s.TeamvaultConnector.Password(s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault password failed")
	}
	return []byte(teamvaultPassword), nil
}

func (s *SecretFromTeamvault) Validate(ctx context.Context) error {
	_, err := s.Value()
	return errors.Wrapf(err, "get teamvault secret %s failed", s.TeamvaultKey.String())
}

type SecretValueStatic struct {
	Content []byte
}

func (s *SecretValueStatic) Value() ([]byte, error) {
	return s.Content, nil
}

func (s *SecretValueStatic) Validate(ctx context.Context) error {
	return nil
}

type Secrets map[string]SecretValue

type MountName string

type MountTarget string

type MountReadOnly bool

type MountNfsPath string

type MountNfsServer string

type Mount struct {
	Name     MountName
	Target   MountTarget
	ReadOnly MountReadOnly
}

func (m *Mount) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating mount %s", m.Name)
	if m.Name == "" {
		return errors.New("name missing")
	}
	if m.Target == "" {
		return errors.New("target missing")
	}
	glog.V(4).Infof("mount %s is valid", m.Name)
	return nil
}

type Volume struct {
	Name      MountName
	NfsPath   MountNfsPath
	NfsServer MountNfsServer
	EmptyDir  bool
}

func (m *Volume) GetName() MountName {
	return m.Name
}

func (m *Volume) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating mount %s", m.Name)
	if m.Name == "" {
		return errors.New("name missing")
	}
	if m.NfsPath == "" {
		return errors.New("nfs path missing")
	}
	if m.NfsServer == "" {
		return errors.New("nfs server missing")
	}
	glog.V(4).Infof("mount %s is valid", m.Name)
	return nil
}

type LivenessProbe bool

type ReadinessProbe bool

type ContainerName string

func (c ContainerName) String() string {
	return string(c)
}

type CpuLimit string

func (b CpuLimit) String() string {
	return string(b)
}

type MemoryLimit string

func (b MemoryLimit) String() string {
	return string(b)
}

type CpuRequest string

func (b CpuRequest) String() string {
	return string(b)
}

type MemoryRequest string

func (b MemoryRequest) String() string {
	return string(b)
}

type DeploymentDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Containers   []DeploymentDeployerContainer
	Volumes      []Volume
	HostNetwork  k8s.PodHostNetwork
}

type DeploymentDeployerContainer struct {
	Name          ContainerName
	Args          []k8s.Arg
	Ports         []Port
	Env           []k8s.Env
	CpuLimit      CpuLimit
	MemoryLimit   MemoryLimit
	CpuRequest    CpuRequest
	MemoryRequest MemoryRequest
	Mounts        []Mount
	Image         docker.Image
}

func (d *DeploymentDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: d.Context,
		Data:    d,
	}
}

func (d *DeploymentDeployer) Children() []world.Configuration {
	return d.Requirements
}

func (d *DeploymentDeployer) Validate(ctx context.Context) error {
	if d.Context == "" {
		return errors.New("Context missing")
	}
	if d.Namespace == "" {
		return errors.New("Namespace missing")
	}
	if d.Name == "" {
		return errors.New("Name missing")
	}
	if len(d.Containers) == 0 {
		return errors.New("Containers missing")
	}
	for _, container := range d.Containers {
		if container.Name == "" {
			return errors.New("Name missing")
		}
		if container.CpuLimit == "" {
			return errors.New("CpuLimit missing")
		}
		if container.MemoryLimit == "" {
			return errors.New("MemoryLimit missing")
		}
		if container.CpuRequest == "" {
			return errors.New("CpuRequest missing")
		}
		if container.MemoryRequest == "" {
			return errors.New("MemoryRequest missing")
		}
		if err := container.Image.Validate(ctx); err != nil {
			return errors.Wrap(err, "image invalid")
		}
	}
	return nil
}

func (d *DeploymentDeployer) Data() (interface{}, error) {
	return d.deployment(), nil
}

func (d *DeploymentDeployer) deployment() k8s.Deployment {
	var volumes []k8s.PodVolume
	for _, volume := range d.Volumes {
		podVolume := k8s.PodVolume{
			Name: k8s.PodVolumeName(volume.Name),
		}
		if volume.EmptyDir {
			podVolume.EmptyDir = k8s.EmptyDir{}
		} else {
			podVolume.Nfs = k8s.PodNfs{
				Path:   k8s.PodNfsPath(volume.NfsPath),
				Server: k8s.PodNfsServer(volume.NfsServer),
			}
		}
		volumes = append(volumes, podVolume)
	}
	return k8s.Deployment{
		ApiVersion: "extensions/v1beta1",
		Kind:       "Deployment",
		Metadata: k8s.Metadata{
			Namespace: d.Namespace,
			Name:      d.Name,
			Labels: k8s.Labels{
				"app": d.Name.String(),
			},
		},
		Spec: k8s.DeploymentSpec{
			Replicas:             1,
			RevisionHistoryLimit: 2,
			Selector: k8s.DeploymentSelector{
				MatchLabels: k8s.DeploymentMatchLabels{
					"app": d.Name.String(),
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
						"app": d.Name.String(),
					},
				},
				Spec: k8s.PodSpec{
					Containers:  d.containers(),
					Volumes:     volumes,
					HostNetwork: d.HostNetwork,
				},
			},
		},
	}
}

func (d *DeploymentDeployer) containers() []k8s.PodContainer {
	var result []k8s.PodContainer
	for _, container := range d.Containers {
		result = append(result, d.container(container))
	}
	return result
}

func (d *DeploymentDeployer) container(container DeploymentDeployerContainer) k8s.PodContainer {
	var mounts []k8s.VolumeMount
	for _, mount := range container.Mounts {
		mounts = append(mounts, k8s.VolumeMount{
			MountPath: k8s.VolumeMountPath(mount.Target),
			Name:      k8s.VolumeName(mount.Name),
			ReadOnly:  k8s.VolumeReadOnly(mount.ReadOnly),
		})
	}
	podContainer := k8s.PodContainer{
		Image: k8s.PodImage(container.Image.String()),
		Name:  k8s.PodName(container.Name.String()),
		Resources: k8s.PodResources{
			Requests: k8s.ResourceList{
				"cpu":    container.CpuRequest.String(),
				"memory": container.MemoryRequest.String(),
			},
			Limits: k8s.ResourceList{
				"cpu":    container.CpuLimit.String(),
				"memory": container.MemoryLimit.String(),
			},
		},
		VolumeMounts: mounts,
	}
	for _, port := range container.Ports {
		podContainer.Ports = append(podContainer.Ports, k8s.PodPort{
			Name:          k8s.PodPortName(port.Name),
			ContainerPort: k8s.PodPortContainerPort(port.Port),
			HostPort:      k8s.PodPortHostPort(port.HostPort),
			Protocol:      k8s.PodPortProtocol(port.Protocol),
		})
	}
	for _, arg := range container.Args {
		podContainer.Args = append(podContainer.Args, k8s.Arg(arg))
	}
	podContainer.Env = container.Env
	return podContainer
}
