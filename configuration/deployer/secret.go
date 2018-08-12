package deployer

import (
	"context"
	"encoding/base64"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type SecretDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Secrets      world.Secrets
}

func (i *SecretDeployer) Applier() world.Applier {
	return &k8s.Deployer{
		Context: i.Context,
		Data:    i,
	}
}

func (i *SecretDeployer) Childs() []world.Configuration {
	return i.Requirements
}

func (i *SecretDeployer) Validate(ctx context.Context) error {
	if i.Context == "" {
		return errors.New("Context missing in secret deployer")
	}
	if i.Namespace == "" {
		return errors.New("Namespace missing in secret deployer")
	}
	if i.Namespace == "" {
		return errors.New("Name missing in secret deployer")
	}
	return nil
}

func (i *SecretDeployer) Data() (interface{}, error) {
	return i.secret()
}

func (i *SecretDeployer) secret() (*k8s.Secret, error) {
	secret := &k8s.Secret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Metadata: k8s.Metadata{
			Namespace: i.Namespace,
			Name:      i.Name,
			Labels: k8s.Labels{
				"app": i.Namespace.String(),
			},
		},
		Type: "Opaque",
		Data: k8s.SecretData{},
	}
	for k, v := range i.Secrets {
		value, err := v.Value()
		if err != nil {
			return nil, err
		}
		secret.Data[k] = base64.StdEncoding.EncodeToString(value)
	}
	return secret, nil
}
