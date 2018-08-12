package world

import (
	"context"

	"github.com/bborbe/teamvault-utils/connector"
	"github.com/bborbe/teamvault-utils/model"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate counterfeiter -o mocks/configuration.go --fake-name Configuration . Configuration
type Configuration interface {
	Childs() []Configuration
	Applier() Applier
	Validate(ctx context.Context) error
}

//go:generate counterfeiter -o mocks/applier.go --fake-name Applier . Applier
type Applier interface {
	Satisfied(ctx context.Context) (bool, error)
	Apply(ctx context.Context) error
	Validate(ctx context.Context) error
}

type Domain string

func (d Domain) String() string {
	return string(d)
}

type BuilderType string

func (b BuilderType) String() string {
	return string(b)
}

type Port struct {
	Name     string
	Port     int
	HostPort int
	Protocol string
}

type Arg string

type SecretValue interface {
	Value() ([]byte, error)
	Validate(ctx context.Context) error
}

type SecretFromTeamvault struct {
	TeamvaultConnector connector.Connector
	TeamvaultKey       model.TeamvaultKey
}

func (s *SecretFromTeamvault) Value() ([]byte, error) {
	teamvaultPassword, err := s.TeamvaultConnector.Password(s.TeamvaultKey)
	if err != nil {
		return nil, errors.Wrap(err, "get teamvault password failed")
	}
	return []byte(teamvaultPassword), nil
}

func (s *SecretFromTeamvault) Validate(ctx context.Context) error {
	_, err := s.Value()
	return errors.Wrapf(err, "get teamvault secret %s failed", s.TeamvaultKey.String())
}

type SecretValueStatic struct {
	Content []byte
}

func (s *SecretValueStatic) Value() ([]byte, error) {
	return s.Content, nil
}

func (s *SecretValueStatic) Validate(ctx context.Context) error {
	return nil
}

type Secrets map[string]SecretValue

type CpuLimit string

func (b CpuLimit) String() string {
	return string(b)
}

type MemoryLimit string

func (b MemoryLimit) String() string {
	return string(b)
}

type CpuRequest string

func (b CpuRequest) String() string {
	return string(b)
}

type MemoryRequest string

func (b MemoryRequest) String() string {
	return string(b)
}

type MountName string

type MountTarget string

type MountReadOnly bool

type MountNfsPath string

type MountNfsServer string

type Mount struct {
	Name     MountName
	Target   MountTarget
	ReadOnly MountReadOnly
}

func (m *Mount) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating mount %s", m.Name)
	if m.Name == "" {
		return errors.New("name missing")
	}
	if m.Target == "" {
		return errors.New("target missing")
	}
	glog.V(4).Infof("mount %s is valid", m.Name)
	return nil
}

type Volume struct {
	Name      MountName
	NfsPath   MountNfsPath
	NfsServer MountNfsServer
	EmptyDir  bool
}

func (m *Volume) GetName() MountName {
	return m.Name
}

func (m *Volume) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating mount %s", m.Name)
	if m.Name == "" {
		return errors.New("name missing")
	}
	if m.NfsPath == "" {
		return errors.New("nfs path missing")
	}
	if m.NfsServer == "" {
		return errors.New("nfs server missing")
	}
	glog.V(4).Infof("mount %s is valid", m.Name)
	return nil
}

type LivenessProbe bool

type ReadinessProbe bool

type ContainerName string

func (c ContainerName) String() string {
	return string(c)
}
