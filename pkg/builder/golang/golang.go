package golang

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/builder"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Builder struct {
	Image           world.Image
	Name            world.Name
	SourceDirectory world.SourceDirectory
	GitRepo         world.GitRepo
	Package         world.Package
}

func (b *Builder) Build(ctx context.Context) error {
	glog.V(1).Infof("building golang docker image %s ...", b.Name)
	tmpl, err := template.New("template").Parse(`
FROM golang:1.10 AS build
RUN git clone --branch {{.Tag}} --single-branch --depth 1 {{.GitRepo}} ./src/{{.SourceDirectory}} 
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /{{.Name}} ./src/{{.Package}}

FROM scratch
MAINTAINER Benjamin Borbe <bborbe@rocketnews.de>
COPY --from=build /{{.Name}} /{{.Name}}
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/{{.Name}}"]
`)
	if err != nil {
		return errors.Wrap(err, "parse dockerfile template failed")
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
		Package         world.Package
		Name            world.Name
		SourceDirectory world.SourceDirectory
		GitRepo         world.GitRepo
		Tag             world.Tag
	}{
		Package:         b.Package,
		Name:            b.Name,
		SourceDirectory: b.SourceDirectory,
		GitRepo:         b.GitRepo,
		Tag:             b.GetImage().Tag,
	})
	if err != nil {
		return errors.Wrap(err, "fill dockerfile template failed")
	}
	cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "--rm=true", "--tag", b.Image.String(), "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "build docker image failed")
	}
	glog.V(1).Infof("building golang docker image %s finished", b.Name)
	return nil
}

func (b *Builder) Validate(ctx context.Context) error {
	if err := b.Image.Validate(ctx); err != nil {
		return err
	}
	if b.Name == "" {
		return errors.New("name missing")
	}
	if b.SourceDirectory == "" {
		return errors.New("source directory missing")
	}
	if b.GitRepo == "" {
		return errors.New("git repo missing")
	}
	if b.Package == "" {
		return errors.New("package missing")
	}
	return nil
}

func (b *Builder) GetImage() world.Image {
	return b.Image
}

func (b *Builder) Satisfied(ctx context.Context) (bool, error) {
	return builder.DockerImageExists(ctx, b.Image)
}
