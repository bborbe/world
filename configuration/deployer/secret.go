package deployer

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
)

type SecretDeployer struct {
	Context      world.Context
	Requirements []world.Configuration
	Namespace    world.Namespace
	Secrets      world.Secrets
}

func (i *SecretDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: i.Context,
		Data:    i.secret(),
	}
}

func (i *SecretDeployer) Childs() []world.Configuration {
	return i.Requirements
}

func (i *SecretDeployer) Validate(ctx context.Context) error {
	if i.Context == "" {
		return fmt.Errorf("Context missing")
	}
	if i.Namespace == "" {
		return fmt.Errorf("Namespace missing")
	}
	return nil
}

func (i *SecretDeployer) secret() k8s.Secret {
	secret := k8s.Secret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Metadata: k8s.Metadata{
			Namespace: k8s.NamespaceName(i.Namespace),
			Name:      k8s.Name(i.Namespace),
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
		},
		Type: "Opaque",
		Data: k8s.SecretData{},
	}
	for k, v := range i.Secrets {
		secret.Data[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	return secret
}
