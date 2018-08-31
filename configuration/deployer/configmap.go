package deployer

import (
	"context"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type ConfigMapDeployer struct {
	Context       k8s.Context
	Namespace     k8s.NamespaceName
	Name          k8s.MetadataName
	ConfigMapData k8s.ConfigMapData
	Requirements  []world.Configuration
}

func (d *ConfigMapDeployer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.Context,
		d.Namespace,
		d.Name,
		d.ConfigMapData,
	)
}

func (i *ConfigMapDeployer) Applier() (world.Applier, error) {
	return &k8s.ConfigMapApplier{
		Context:   i.Context,
		ConfigMap: i.configMap(),
	}, nil
}

func (i *ConfigMapDeployer) Children() []world.Configuration {
	return i.Requirements
}

func (i *ConfigMapDeployer) configMap() k8s.ConfigMap {
	return k8s.ConfigMap{
		ApiVersion: "v1",
		Kind:       "ConfigMap",
		Metadata: k8s.Metadata{
			Namespace: i.Namespace,
			Name:      i.Name,
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
		},
		Data: i.ConfigMapData,
	}
}
