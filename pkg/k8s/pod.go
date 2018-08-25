package k8s

import (
	"context"

	"github.com/pkg/errors"
)

type PodSpec struct {
	Containers  []Container    `yaml:"containers"`
	Volumes     []PodVolume    `yaml:"volumes,omitempty"`
	HostNetwork PodHostNetwork `yaml:"hostNetwork,omitempty"`
	DnsPolicy   PodDnsPolicy   `yaml:"dnsPolicy,omitempty"`
}

type PodVolumeName string

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
	Key  PodSecretItemKey  `yaml:"key"`
	Path PodSecretItemPath `yaml:"path"`
}

type PodSecretItemKey string

type PodSecretItemPath string

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
	Name      PodVolumeName      `yaml:"name"`
	Nfs       PodVolumeNfs       `yaml:"nfs,omitempty"`
	ConfigMap PodVolumeConfigMap `yaml:"configMap,omitempty"`
	Secret    PodVolumeSecret    `yaml:"secret,omitempty"`
	EmptyDir  *PodVolumeEmptyDir `yaml:"emptyDir,omitempty"`
}

type PodDnsPolicy string

type PodHostNetwork bool

type ValueFrom struct {
	SecretKeyRef SecretKeyRef `yaml:"secretKeyRef"`
}

type SecretKeyRef struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}
