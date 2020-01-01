// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deployer

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"

	teamvault "github.com/bborbe/teamvault-utils"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
)

type SecretValue interface {
	Value(ctx context.Context) ([]byte, error)
	Validate(ctx context.Context) error
}

type SecretFromFile struct {
	Path string
}

func (s SecretFromFile) Validate(ctx context.Context) error {
	if s.Path == "" {
		return errors.New("Path missing")
	}
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return errors.Errorf("file not found %s", s.Path)
	}
	return nil
}

func (s *SecretFromFile) Value(ctx context.Context) ([]byte, error) {
	return ioutil.ReadFile(s.Path)
}

type SecretFromTeamvaultUser struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
}

func (s SecretFromTeamvaultUser) Validate(ctx context.Context) error {
	if s.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	if s.TeamvaultKey == "" {
		return errors.New("TeamvaultKey missing")
	}
	return nil
}

func (s SecretFromTeamvaultUser) Value(ctx context.Context) ([]byte, error) {
	teamvaultUsername, err := s.TeamvaultConnector.User(ctx, s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault username failed")
	}
	return []byte(teamvaultUsername), nil
}

type SecretFromTeamvaultFile struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
}

func (s SecretFromTeamvaultFile) Validate(ctx context.Context) error {
	if s.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	if s.TeamvaultKey == "" {
		return errors.New("TeamvaultKey missing")
	}
	return nil
}

func (s SecretFromTeamvaultFile) Value(ctx context.Context) ([]byte, error) {
	teamvaultFile, err := s.TeamvaultConnector.File(ctx, s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault filename failed")
	}
	return teamvaultFile.Content()
}

type SecretFromTeamvaultHtpasswd struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
}

func (s SecretFromTeamvaultHtpasswd) Validate(ctx context.Context) error {
	if s.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	if s.TeamvaultKey == "" {
		return errors.New("TeamvaultKey missing")
	}
	return nil
}

func (s SecretFromTeamvaultHtpasswd) Value(ctx context.Context) ([]byte, error) {
	htpasswd := teamvault.Htpasswd{Connector: s.TeamvaultConnector}
	bytes, err := htpasswd.Generate(ctx, s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault htpasswd failed")
	}
	return bytes, nil
}

type SecretFromTeamvaultPassword struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
}

func (s SecretFromTeamvaultPassword) Validate(ctx context.Context) error {
	if s.TeamvaultConnector == nil {
		return errors.New("TeamvaultConnector missing")
	}
	if s.TeamvaultKey == "" {
		return errors.New("TeamvaultKey missing")
	}
	return nil
}

func (s SecretFromTeamvaultPassword) Value(ctx context.Context) ([]byte, error) {
	teamvaultPassword, err := s.TeamvaultConnector.Password(ctx, s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault password failed")
	}
	return []byte(teamvaultPassword), nil
}

type SecretValueStatic []byte

func (s SecretValueStatic) Value(ctx context.Context) ([]byte, error) {
	return s, nil
}

func (s SecretValueStatic) Validate(ctx context.Context) error {
	return nil
}

type Secrets map[string]SecretValue

func (s Secrets) Validate(ctx context.Context) error {
	for k, v := range s {
		if k == "" {
			return errors.New("secret has no name")
		}
		if err := v.Validate(ctx); err != nil {
			return errors.Wrapf(err, "value of secret %s invalid", k)
		}
	}
	return nil
}

type SecretApplier struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	Secrets      Secrets
	Requirements []world.Configuration
}

func (s *SecretApplier) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *SecretApplier) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		s.Context,
		s.Namespace,
		s.Name,
		s.Secrets,
	)
}

func (s *SecretApplier) Apply(ctx context.Context) error {
	secret, err := s.secret(ctx)
	if err != nil {
		return err
	}
	applier := &k8s.SecretApplier{
		Context: s.Context,
		Secret:  *secret,
	}
	return applier.Apply(ctx)
}

func (s *SecretApplier) Children() []world.Configuration {
	return s.Requirements
}

func (s *SecretApplier) secret(ctx context.Context) (*k8s.Secret, error) {
	secret := &k8s.Secret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Metadata: k8s.Metadata{
			Namespace: s.Namespace,
			Name:      s.Name,
			Labels: k8s.Labels{
				"app": s.Namespace.String(),
			},
		},
		Type: "Opaque",
		Data: k8s.SecretData{},
	}
	for k, v := range s.Secrets {
		value, err := v.Value(ctx)
		if err != nil {
			return nil, err
		}
		secret.Data[k] = base64.StdEncoding.EncodeToString(value)
	}
	return secret, nil
}
