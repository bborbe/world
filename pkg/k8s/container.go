// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/validation"
)

type Image string

type ContainerName string

type CpuLimit string

func (c CpuLimit) Validate(ctx context.Context) error {
	if c == "" {
		return errors.New("CpuLimit empty")
	}
	return nil
}

type MemoryLimit string

func (m MemoryLimit) Validate(ctx context.Context) error {
	if m == "" {
		return errors.New("MemoryLimit empty")
	}
	return nil
}

type ContainerResource struct {
	Cpu    CpuLimit    `yaml:"cpu"`
	Memory MemoryLimit `yaml:"memory"`
}

func (c ContainerResource) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Cpu,
		c.Memory,
	)
}

type Resources struct {
	Limits   ContainerResource `yaml:"limits"`
	Requests ContainerResource `yaml:"requests"`
}

func (r Resources) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		r.Limits,
		r.Requests,
	)
}

type ContainerPorts []ContainerPort

func (c ContainerPorts) Validate(ctx context.Context) error {
	containerPorts := make(map[string]struct{})
	for _, port := range c {
		protocol := strings.ToUpper(port.Protocol.String())
		if protocol == "" {
			protocol = "TCP"
		}
		key := fmt.Sprint(protocol, port.ContainerPort)
		_, ok := containerPorts[key]
		if ok {
			return errors.Errorf("duplicate container port %s %s", port.Protocol, port.ContainerPort)
		}
		containerPorts[key] = struct{}{}
	}
	return nil
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

type ImagePullPolicy string

type Container struct {
	Name            ContainerName    `yaml:"name"`
	Image           Image            `yaml:"image"`
	Command         []Command        `yaml:"command,omitempty"`
	Args            []Arg            `yaml:"args,omitempty"`
	Env             []Env            `yaml:"env,omitempty"`
	Ports           ContainerPorts   `yaml:"ports,omitempty"`
	Resources       Resources        `yaml:"resources,omitempty"`
	VolumeMounts    []ContainerMount `yaml:"volumeMounts,omitempty"`
	ReadinessProbe  Probe            `yaml:"readinessProbe,omitempty"`
	LivenessProbe   Probe            `yaml:"livenessProbe,omitempty"`
	SecurityContext SecurityContext  `yaml:"securityContext,omitempty"`
	ImagePullPolicy ImagePullPolicy  `yaml:"imagePullPolicy,omitempty"`
}

func (c Container) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		c.Resources,
		c.Ports,
	)
}

type SecurityContext struct {
	AllowPrivilegeEscalation bool                        `yaml:"allowPrivilegeEscalation,omitempty"`
	ReadOnlyRootFilesystem   bool                        `yaml:"readOnlyRootFilesystem,omitempty"`
	Privileged               bool                        `yaml:"privileged,omitempty"`
	RunAsUser                int                         `yaml:"runAsUser,omitempty"`
	FsGroup                  int                         `yaml:"fsGroup,omitempty"`
	Capabilities             SecurityContextCapabilities `yaml:"capabilities,omitempty"`
}

type SecurityContextCapabilities map[string][]string

type Probe struct {
	Exec                Exec      `yaml:"exec,omitempty"`
	HttpGet             HttpGet   `yaml:"httpGet,omitempty"`
	TcpSocket           TcpSocket `yaml:"tcpSocket,omitempty"`
	InitialDelaySeconds int       `yaml:"initialDelaySeconds,omitempty"`
	SuccessThreshold    int       `yaml:"successThreshold,omitempty"`
	FailureThreshold    int       `yaml:"failureThreshold,omitempty"`
	TimeoutSeconds      int       `yaml:"timeoutSeconds,omitempty"`
	PeriodSeconds       int       `yaml:"periodSeconds,omitempty"`
}

type HttpGet struct {
	Host   string     `yaml:"host,omitempty"`
	Path   string     `yaml:"path,omitempty"`
	Port   PortNumber `yaml:"port,omitempty"`
	Scheme string     `yaml:"scheme,omitempty"`
}

type TcpSocket struct {
	Port PortNumber `yaml:"port,omitempty"`
}

type Exec struct {
	Command []Command `yaml:"command,omitempty"`
}
