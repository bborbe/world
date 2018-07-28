package world

import (
	"context"

	"github.com/bborbe/run"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate counterfeiter -o mocks/applier.go --fake-name Applier . Applier
type Applier interface {
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	Apply(ctx context.Context) error
	Required() Applier
}

type Appliers []Applier

func (a Appliers) Validate(ctx context.Context) error {
	for _, applier := range a {
		if err := applier.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a Appliers) Satisfied(ctx context.Context) (bool, error) {
	for _, applier := range a {
		ok, err := applier.Satisfied(ctx)
		if err != nil {
			return ok, err
		}
		if !ok {
			return ok, err
		}
	}
	return true, nil
}

func (a Appliers) Required() Applier {
	return nil
}

func (a Appliers) Apply(ctx context.Context) error {
	glog.V(4).Infof("apply app ...")
	defer glog.V(4).Infof("apply app finished")
	var list []run.RunFunc
	for _, applier := range a {
		list = append(list, func(ctx context.Context) error {
			ok, err := applier.Satisfied(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			if applier.Required() != nil {
				if err := applier.Required().Apply(ctx); err != nil {
					return errors.Wrap(err, "apply requirements failed")
				}
			}
			return applier.Apply(ctx)
		})
	}
	return run.Sequential(ctx, list...)
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
