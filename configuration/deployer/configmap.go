package deployer

import (
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type ConfigMapDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	ConfigMap    k8s.ConfigMapData
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
		Data: i.ConfigMap,
	}
}
