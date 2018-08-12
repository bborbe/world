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
		return errors.New("name missing in golang builder")
	}
	if g.SourceDirectory == "" {
		return errors.New("source directory missing in golang builder")
	}
	if g.GitRepo == "" {
		return errors.New("git repo missing in golang builder")
	}
	if g.Package == "" {
		return errors.New("package missing in golang builder")
	}
	return nil
}
