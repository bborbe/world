package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type SecretApplier struct {
	Context Context
	Secret  Secret
}

func (s *SecretApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *SecretApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Secret,
	}
	return deployer.Apply(ctx)
}

func (s *SecretApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Secret.Validate(ctx)
}

type Secret struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
	Type       SecretType `yaml:"type"`
	Data       SecretData `yaml:"data"`
}

func (s *Secret) Validate(ctx context.Context) error {
	if s.ApiVersion != "v1" {
		return errors.New("invalid ApiVersion")
	}
	if s.Kind != "Secret" {
		return errors.New("invalid Kind")
	}
	return nil
}

func (s Secret) String() string {
	return fmt.Sprintf("%s/%s to %s", s.Kind, s.Metadata.Name, s.Metadata.Namespace)
}

type SecretType string

type SecretData map[string]string
