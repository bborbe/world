package k8s

import (
	"context"

	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type PodTemplate struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     PodSpec  `yaml:"spec"`
}

func (p PodTemplate) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		p.Spec,
	)
}

type Toleration struct {
	Key      string `yaml:"key,omitempty"`
	Effect   string `yaml:"effect,omitempty"`
	Operator string `yaml:"operator,omitempty"`
}

type PodSpec struct {
	Tolerations     []Toleration    `yaml:"tolerations,omitempty"`
	Containers      []Container     `yaml:"containers"`
	Volumes         []PodVolume     `yaml:"volumes,omitempty"`
	HostNetwork     PodHostNetwork  `yaml:"hostNetwork,omitempty"`
	HostPid         PodHostPID      `yaml:"hostPID,omitempty"`
	DnsPolicy       PodDnsPolicy    `yaml:"dnsPolicy,omitempty"`
	SecurityContext SecurityContext `yaml:"securityContext,omitempty"`
}

func (c PodSpec) Validate(ctx context.Context) error {
	if len(c.Containers) == 0 {
		return errors.New("Containers empty")
	}
	for _, container := range c.Containers {
		if err := container.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type PodNfsPath string

func (d PodNfsPath) Validate(ctx context.Context) error {
	if d == "" {
		return errors.New("PodNfsPath empty")
	}
	return nil
}

type PodNfsServer string

func (d PodNfsServer) Validate(ctx context.Context) error {
	if d == "" {
		return errors.New("PodNfsServer empty")
	}
	return nil
}

type PodHostPath string

func (p PodHostPath) Validate(ctx context.Context) error {
	if p == "" {
		return errors.New("PodHostPath missing")
	}
	return nil
}

type PodVolumeHost struct {
	Path PodHostPath `yaml:"path"`
}

type PodVolumeNfs struct {
	Path   PodNfsPath   `yaml:"path"`
	Server PodNfsServer `yaml:"server"`
}

type PodVolumeSecret struct {
	Name  PodSecretName   `yaml:"secretName"`
	Items []PodSecretItem `yaml:"items"`
}

type PodSecretName string

type PodSecretItem struct {
	Key  PodSecretItemKey  `yaml:"key,omitempty"`
	Mode PodSecretItemMode `yaml:"mode,omitempty"`
	Path PodSecretItemPath `yaml:"path,omitempty"`
}

type PodSecretItemKey string

type PodSecretItemPath string

type PodSecretItemMode int

type PodVolumeConfigMap struct {
	Name  PodConfigMapName   `yaml:"name"`
	Items []PodConfigMapItem `yaml:"items"`
}

type PodConfigMapName string

type PodConfigMapItem struct {
	Key  PodConfigMapItemKey  `yaml:"key"`
	Path PodConfigMapItemPath `yaml:"path"`
}

type PodConfigMapItemKey string

type PodConfigMapItemPath string

type PodVolumeEmptyDir struct{}

type PodVolume struct {
	Name      MountName          `yaml:"name"`
	Nfs       PodVolumeNfs       `yaml:"nfs,omitempty"`
	Host      PodVolumeHost      `yaml:"hostPath,omitempty"`
	ConfigMap PodVolumeConfigMap `yaml:"configMap,omitempty"`
	Secret    PodVolumeSecret    `yaml:"secret,omitempty"`
	EmptyDir  *PodVolumeEmptyDir `yaml:"emptyDir,omitempty"`
}

type PodDnsPolicy string

type PodHostNetwork bool

type PodHostPID bool

type FieldRef struct {
	FieldPath string `yaml:"fieldPath"`
}

type ValueFrom struct {
	SecretKeyRef SecretKeyRef `yaml:"secretKeyRef,omitempty"`
	FieldRef     FieldRef     `yaml:"fieldRef,omitempty"`
}

type SecretKeyRef struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}
