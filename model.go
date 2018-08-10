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

type Registry string

func (r Registry) String() string {
	return string(r)
}

type Repository string

func (i Repository) String() string {
	return string(i)
}

type Tag string

func (v Tag) String() string {
	return string(v)
}

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type GitRepo string

func (g GitRepo) String() string {
	return string(g)
}

type GitBranch string

func (g GitBranch) String() string {
	return string(g)
}

type Package string

func (p Package) String() string {
	return string(p)
}

type Name string

func (n Name) String() string {
	return string(n)
}

type Namespace string

func (n Namespace) String() string {
	return string(n)
}

type HostNetwork bool

type Domain string

func (d Domain) String() string {
	return string(d)
}

type BuilderType string

func (b BuilderType) String() string {
	return string(b)
}

type Image struct {
	Repository Repository
	Registry   Registry
	Tag        Tag
}

func (i Image) String() string {
	return i.Registry.String() + "/" + i.Repository.String() + ":" + i.Tag.String()
}

func (b *Image) Validate(ctx context.Context) error {
	if b.Tag == "" {
		return errors.New("tag missing")
	}
	if b.Registry == "" {
		return errors.New("tag missing")
	}
	if b.Repository == "" {
		return errors.New("tag missing")
	}
	return nil
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

type SecretValueStatic struct {
	Content []byte
}

func (s *SecretValueStatic) Value() ([]byte, error) {
	return s.Content, nil
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

type Context string

func (c Context) String() string {
	return string(c)
}

type ContainerName string

func (c ContainerName) String() string {
	return string(c)
}

type BuildArgs map[string]string
