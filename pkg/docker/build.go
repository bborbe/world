package docker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"github.com/bborbe/run"
	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func BuilderForApp(app world.App) (*Builder, error) {
	buildInstruction, err := world.BuildInstructionForName(app.Name)
	if err != nil {
		return nil, errors.Wrap(err, "find build instructions failed")
	}
	return New(buildInstruction, app.Version), nil
}

func New(buildInstruction *world.BuildInstruction, version world.Version) *Builder {
	return &Builder{
		Registry:        buildInstruction.Registry,
		Image:           buildInstruction.Image,
		Version:         version,
		Name:            buildInstruction.Name,
		SourceDirectory: buildInstruction.SourceDirectory,
		GitRepo:         buildInstruction.GitRepo,
		Package:         buildInstruction.Package,
	}
}

type Builder struct {
	Registry        world.Registry
	Image           world.Image
	Version         world.Version
	Name            world.Name
	SourceDirectory world.SourceDirectory
	GitRepo         world.GitRepo
	Package         world.Package
}

func (b *Builder) Build(ctx context.Context) error {
	glog.V(1).Infof("building %s ...", b.Name)
	return errors.Wrap(run.Sequential(
		ctx,
		b.buildDockerImage,
		b.uploadDockerImage,
	), "build failed")
}

func (b *Builder) buildDockerImage(ctx context.Context) error {
	glog.V(2).Infof("run docker build ...")
	tmpl, err := template.New("template").Parse(`
FROM golang:1.10 AS build
RUN git clone --branch {{.Version}} --single-branch --depth 1 {{.GitRepo}} ./src/{{.SourceDirectory}} 
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
		Package         string
		Name            string
		Version         string
		SourceDirectory string
		GitRepo         string
	}{
		Package:         b.Package.String(),
		Name:            b.Name.String(),
		Version:         b.Version.String(),
		SourceDirectory: b.SourceDirectory.String(),
		GitRepo:         b.GitRepo.String(),
	})
	if err != nil {
		return errors.Wrap(err, "fill dockerfile template failed")
	}
	cmd := exec.CommandContext(ctx, "docker", "build", "--no-cache", "--rm=true", "--tag", b.Registry.String()+"/"+b.Image.String()+":"+b.Version.String(), "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "build docker image failed")
}

func (b *Builder) uploadDockerImage(ctx context.Context) error {
	glog.V(2).Infof("run docker build ...")
	cmd := exec.CommandContext(ctx, "docker", "push", b.Registry.String()+"/"+b.Image.String()+":"+b.Version.String())
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "upload docker image failed")
}
