package world

import (
	"context"
	"fmt"

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

type Context string

func (c Context) String() string {
	return string(c)
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

type Builder interface {
	Build(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	GetImage() Image
}

type Uploader interface {
	Upload(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	GetBuilder() Builder
}

type Deployer interface {
	Deploy(ctx context.Context) error
	Validate(ctx context.Context) error
	Satisfied(ctx context.Context) (bool, error)
	GetUploader() Uploader
}

type App struct {
	Name     Name
	Deployer Deployer
}

func (a *App) Validate(ctx context.Context) error {
	glog.V(4).Infof("validating app %s", a.Name)
	if a.Name == "" {
		return errors.New("name missing")
	}
	if a.Deployer == nil {
		return fmt.Errorf("%s has no builder", a.Name)
	}
	if err := a.Deployer.Validate(ctx); err != nil {
		return err
	}
	glog.V(4).Infof("app %s is valid", a.Name)
	return nil
}

type Apps []App

func (a Apps) Validate(ctx context.Context) error {
	for _, app := range a {
		if err := app.Validate(ctx); err != nil {
			return errors.Wrap(err, "validate failed")
		}
	}
	return nil
}

func (a Apps) WithName(name Name) (*App, error) {
	for _, app := range a {
		if app.Name == name {
			return &app, nil
		}
	}
	return nil, fmt.Errorf("no app with name %s found", name.String())
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

type Arg string

type Env map[string]string
