package golang

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"github.com/bborbe/world"
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
	glog.V(1).Infof("building %s ...", b.Name)
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
		Package         string
		Name            string
		SourceDirectory string
		GitRepo         string
	}{
		Package:         b.Package.String(),
		Name:            b.Name.String(),
		SourceDirectory: b.SourceDirectory.String(),
		GitRepo:         b.GitRepo.String(),
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
	return errors.Wrap(cmd.Run(), "build docker image failed")
}
