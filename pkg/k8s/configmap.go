package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type ConfigMapApplier struct {
	Context   Context
	ConfigMap ConfigMap
}

func (s *ConfigMapApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *ConfigMapApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.ConfigMap,
	}
	return deployer.Apply(ctx)
}

func (s *ConfigMapApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.ConfigMap.Validate(ctx)
}

type ConfigMap struct {
	ApiVersion ApiVersion    `yaml:"apiVersion"`
	Kind       Kind          `yaml:"kind"`
	Metadata   Metadata      `yaml:"metadata"`
	Data       ConfigMapData `yaml:"data"`
}

func (c *ConfigMap) Validate(ctx context.Context) error {
	return nil
}

func (c ConfigMap) String() string {
	return fmt.Sprintf("%s/%s to %s", c.Kind, c.Metadata.Name, c.Metadata.Namespace)
}

type ConfigMapType string

type ConfigMapData map[string]string

func (d ConfigMapData) Validate(ctx context.Context) error {
	for k, _ := range d {
		if k == "" {
			return errors.New("Config has no name")
		}
	}
	return nil
}