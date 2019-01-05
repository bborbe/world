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
	Value() ([]byte, error)
	Validate(ctx context.Context) error
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

func (s *SecretFromTeamvaultUser) Value() ([]byte, error) {
	teamvaultUsername, err := s.TeamvaultConnector.User(s.TeamvaultKey)
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

func (s *SecretFromTeamvaultFile) Value() ([]byte, error) {
	teamvaultFile, err := s.TeamvaultConnector.File(s.TeamvaultKey)
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

func (s *SecretFromTeamvaultHtpasswd) Value() ([]byte, error) {
	htpasswd := teamvault.Htpasswd{Connector: s.TeamvaultConnector}
	bytes, err := htpasswd.Generate(s.TeamvaultKey)
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

func (s *SecretFromTeamvaultPassword) Value() ([]byte, error) {
	teamvaultPassword, err := s.TeamvaultConnector.Password(s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault password failed")
	}
	return []byte(teamvaultPassword), nil
}

type SecretValueStatic struct {
	Content []byte
}

func (s *SecretValueStatic) Value() ([]byte, error) {
	return s.Content, nil
}

func (s SecretValueStatic) Validate(ctx context.Context) error {
	return nil
}

type Secrets map[string]SecretValue

func (w Secrets) Validate(ctx context.Context) error {
	for k, v := range w {
		if k == "" {
			return errors.New("secret has no name")
		}
		if err := v.Validate(ctx); err != nil {
			return errors.Wrapf(err, "value of secret %s invalid", k)
		}
	}
	return nil
}

type SecretDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.MetadataName
	Secrets      Secrets
	Requirements []world.Configuration
}

func (w *SecretDeployer) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		w.Context,
		w.Namespace,
		w.Name,
		w.Secrets,
	)
}

func (i *SecretDeployer) Applier() (world.Applier, error) {
	secret, err := i.secret()
	if err != nil {
		return nil, err
	}
	return &k8s.SecretApplier{
		Context: i.Context,
		Secret:  *secret,
	}, nil
}

func (i *SecretDeployer) Children() []world.Configuration {
	return i.Requirements
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

func (s *SecretFromFile) Value() ([]byte, error) {
	return ioutil.ReadFile(s.Path)
}
