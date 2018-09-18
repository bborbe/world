package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/bborbe/world/pkg/docker"

	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Docker struct {
	SSH     ssh.SSH
	Name    remote.ServiceName
	Ports   []int
	Volumes []string
	Image   docker.Image
	Command string
	Args    []string
}

func (d *Docker) Children() []world.Configuration {
	return []world.Configuration{
		&DockerEngine{
			SSH: d.SSH,
		},
		&Service{
			SSH:     d.SSH,
			Name:    d.Name,
			Content: remote.ServiceContent(d.SystemdServiceContent()),
		},
	}
}

func (d *Docker) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *Docker) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}

func (d Docker) SystemdServiceContent() string {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "[Unit]\n")
	fmt.Fprintf(b, "Description=%s\n", d.Name)
	fmt.Fprintf(b, "Requires=docker.service\n")
	fmt.Fprintf(b, "After=docker.service\n")
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, "[Service]\n")
	fmt.Fprintf(b, "Restart=always\n")
	fmt.Fprintf(b, "RestartSec=20s\n")
	fmt.Fprintf(b, "EnvironmentFile=/etc/environment\n")
	fmt.Fprintf(b, "TimeoutStartSec=0\n")
	fmt.Fprintf(b, "ExecStartPre=-/usr/bin/docker kill %s\n", d.Name)
	fmt.Fprintf(b, "ExecStartPre=-/usr/bin/docker rm %s\n", d.Name)
	fmt.Fprintf(b, "ExecStart=/usr/bin/docker run \\\n")
	for _, port := range d.Ports {
		fmt.Fprintf(b, "-p %d:%d \\\n", port, port)
	}
	for _, volume := range d.Volumes {
		fmt.Fprintf(b, "--volume=%s:%s \\\n", volume, volume)
	}
	fmt.Fprintf(b, "--name %s \\\n", d.Name)
	fmt.Fprintf(b, "%s \\\n", d.Image.String())
	fmt.Fprintf(b, "%s \\\n", d.Command)
	fmt.Fprint(b, strings.Join(d.Args, " \\\n"))
	fmt.Fprintf(b, "\n\n")
	fmt.Fprintf(b, "ExecStop=/usr/bin/docker stop %s\n", d.Name)
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, "[Install]\n")
	fmt.Fprintf(b, "WantedBy=multi-user.target\n")
	return b.String()
}
