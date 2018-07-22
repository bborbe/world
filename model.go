package world

import (
	"context"
	"fmt"
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
}

type Uploader interface {
	Upload(ctx context.Context) error
}

type Deployer interface {
	Deploy(ctx context.Context) error
}

type App struct {
	Name     Name
	Builder  Builder
	Uploader Uploader
	Deployer Deployer
}

type Apps []App

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

func (i *Image) String() string {
	return i.Registry.String() + "/" + i.Repository.String() + ":" + i.Tag.String()
}

type Port int

type Arg string

type Env map[string]string
