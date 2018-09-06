package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type Image string

type ContainerName string

type CpuLimit string

type MemoryLimit string

type ContainerResource struct {
	Cpu    string      `yaml:"cpu"`
	Memory MemoryLimit `yaml:"memory"`
}

type Resources struct {
	Limits   ContainerResource `yaml:"limits"`
	Requests ContainerResource `yaml:"requests"`
}

type ContainerPort struct {
	ContainerPort PortNumber   `yaml:"containerPort,omitempty"`
	HostPort      PortNumber   `yaml:"hostPort,omitempty"`
	Name          PortName     `yaml:"name,omitempty"`
	Protocol      PortProtocol `yaml:"protocol,omitempty"`
}

type ContainerMountPath string

type MountName string

func (m MountName) Validate(ctx context.Context) error {
	if m == "" {
		return errors.New("MountName empty")
	}
	return nil
}

type ContainerMountReadOnly bool

type ContainerMount struct {
	Path     ContainerMountPath     `yaml:"mountPath"`
	Name     MountName              `yaml:"name"`
	ReadOnly ContainerMountReadOnly `yaml:"readOnly"`
}

type Arg string

type Command string

type Env struct {
	Name      string    `yaml:"name"`
	Value     string    `yaml:"value,omitempty"`
	ValueFrom ValueFrom `yaml:"valueFrom,omitempty"`
}

type Container struct {
	Name            ContainerName    `yaml:"name"`
	Image           Image            `yaml:"image"`
	Command         []Command        `yaml:"command,omitempty"`
	Args            []Arg            `yaml:"args,omitempty"`
	Env             []Env            `yaml:"env,omitempty"`
	Ports           []ContainerPort  `yaml:"ports,omitempty"`
	Resources       Resources        `yaml:"resources,omitempty"`
	VolumeMounts    []ContainerMount `yaml:"volumeMounts,omitempty"`
	ReadinessProbe  Probe            `yaml:"readinessProbe,omitempty"`
	LivenessProbe   Probe            `yaml:"livenessProbe,omitempty"`
	SecurityContext SecurityContext  `yaml:"securityContext,omitempty"`
}

type SecurityContext struct {
	Privileged bool `yaml:"privileged,omitempty"`
}

type Probe struct {
	HttpGet             HttpGet   `yaml:"httpGet,omitempty"`
	TcpSocket           TcpSocket `yaml:"tcpSocket,omitempty"`
	InitialDelaySeconds int       `yaml:"initialDelaySeconds,omitempty"`
	SuccessThreshold    int       `yaml:"successThreshold,omitempty"`
	FailureThreshold    int       `yaml:"failureThreshold,omitempty"`
	TimeoutSeconds      int       `yaml:"timeoutSeconds,omitempty"`
	PeriodSeconds       int       `yaml:"periodSeconds,omitempty"`
}

type HttpGet struct {
	Path   string     `yaml:"path,omitempty"`
	Port   PortNumber `yaml:"port,omitempty"`
	Scheme string     `yaml:"scheme,omitempty"`
}

type TcpSocket struct {
	Port PortNumber `yaml:"port,omitempty"`
}
