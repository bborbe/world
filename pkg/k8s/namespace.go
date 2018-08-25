package k8s

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type NamespaceName string

func (n NamespaceName) String() string {
	return string(n)
}

func (a NamespaceName) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("NamespaceName missing")
	}
	return nil
}

type NamespaceApplier struct {
	Context   Context
	Namespace Namespace
}

func (s *NamespaceApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *NamespaceApplier) Apply(ctx context.Context) error {
	deployer := &Deployer{
		Context: s.Context,
		Data:    s.Namespace,
	}
	return deployer.Apply(ctx)
}

func (s *NamespaceApplier) Validate(ctx context.Context) error {
	if s.Context == "" {
		return errors.New("context missing")
	}
	return s.Namespace.Validate(ctx)
}

type Namespace struct {
	ApiVersion ApiVersion `yaml:"apiVersion"`
	Kind       Kind       `yaml:"kind"`
	Metadata   Metadata   `yaml:"metadata"`
}

func (s *Namespace) Validate(ctx context.Context) error {
	return validation.Validate(ctx,
		s.ApiVersion,
		s.Kind,
		s.Metadata,
	)
}

func (n Namespace) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Metadata.Name)
}
