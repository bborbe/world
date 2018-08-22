package deployer

import (
	"encoding/base64"

	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/pkg/errors"
)

type SecretValue interface {
	Value() ([]byte, error)
}

type SecretFromTeamvaultUser struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
}

func (s *SecretFromTeamvaultUser) Value() ([]byte, error) {
	teamvaultUsername, err := s.TeamvaultConnector.User(s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault username failed")
	}
	return []byte(teamvaultUsername), nil
}

type SecretFromTeamvaultHtpasswd struct {
	TeamvaultConnector teamvault.Connector
	TeamvaultKey       teamvault.Key
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

type Secrets map[string]SecretValue

type SecretDeployer struct {
	Context      k8s.Context
	Namespace    k8s.NamespaceName
	Name         k8s.Name
	Requirements []world.Configuration
	Secrets      Secrets
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
