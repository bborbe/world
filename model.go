package world

import (
	"context"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

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

type Domain string

func (d Domain) String() string {
	return string(d)
}

type BuilderType string

func (b BuilderType) String() string {
	return string(b)
}

type Apps []App

//go:generate counterfeiter -o mocks/app.go --fake-name App . App
type App interface {
	Apply(ctx context.Context) error
	Validate(ctx context.Context) error
}

//go:generate counterfeiter -o mocks/builder.go --fake-name Builder . Builder
type Builder interface {
	Build(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
}

//go:generate counterfeiter -o mocks/uploader.go --fake-name Uploader . Uploader
type Uploader interface {
	Upload(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
}

//go:generate counterfeiter -o mocks/deployer.go --fake-name Deployer . Deployer
type Deployer interface {
	Deploy(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
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

type Port int

func (p Port) Int() int {
	return int(p)
}

type HostPort int

type Arg string

type Env map[string]string

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
	Name      MountName
	Target    MountTarget
	ReadOnly  MountReadOnly
	NfsPath   MountNfsPath
	NfsServer MountNfsServer
}

func (m *Mount) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating mount %s", m.Name)
	if m.Name == "" {
		return errors.New("name missing")
	}
	if m.Target == "" {
		return errors.New("target missing")
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
