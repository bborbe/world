// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Name string

func (n Name) String() string {
	return string(n)
}

type Package string

func (p Package) String() string {
	return string(p)
}

type GolangBuilder struct {
	Image           Image
	Name            Name
	SourceDirectory SourceDirectory
	GitRepo         GitRepo
	Package         Package
}

func (g *GolangBuilder) Apply(ctx context.Context) error {
	glog.V(1).Infof("building golang docker image %s ...", g.Name)
	tmpl, err := template.New("template").Parse(`
FROM golang:1.13.9 AS build
RUN git clone --branch {{.Tag}} --single-branch --depth 1 {{.GitRepo}} ./src/{{.SourceDirectory}} 
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /{{.Name}} ./src/{{.Package}}

FROM alpine:3.9 as alpine
RUN apk --no-cache add ca-certificates

FROM scratch
MAINTAINER Benjamin Borbe <bborbe@rocketnews.de>
COPY --from=build /{{.Name}} /{{.Name}}
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/{{.Name}}"]
`)
	if err != nil {
		return errors.Wrap(err, "parse dockerfile template failed")
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
		Package         Package
		Name            Name
		SourceDirectory SourceDirectory
		GitRepo         GitRepo
		Tag             Tag
	}{
		Package:         g.Package,
		Name:            g.Name,
		SourceDirectory: g.SourceDirectory,
		GitRepo:         g.GitRepo,
		Tag:             g.Image.Tag,
	})
	if err != nil {
		return errors.Wrap(err, "fill dockerfile template failed")
	}
	cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "--rm=true", "--tag", g.Image.String(), "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "build docker image failed")
	}
	glog.V(1).Infof("building golang docker image %s finished", g.Name)
	return nil
}

func (g *GolangBuilder) Satisfied(ctx context.Context) (bool, error) {
	return ImageExists(ctx, g.Image)
}

func (g *GolangBuilder) Validate(ctx context.Context) error {
	glog.V(4).Infof("validate golang builder ...")
	if err := g.Image.Validate(ctx); err != nil {
		return errors.Wrap(err, "validate golang builder failed")
	}
	if g.Name == "" {
		return errors.New("name missing")
	}
	if g.SourceDirectory == "" {
		return errors.New("source directory missing")
	}
	if g.GitRepo == "" {
		return errors.New("git repo missing")
	}
	if g.Package == "" {
		return errors.New("package missing")
	}
	return nil
}
