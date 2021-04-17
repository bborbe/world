// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/network"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Memory int

func (m Memory) Validate(ctx context.Context) error {
	if m <= 0 {
		return errors.New("Memory empty")
	}
	return nil
}

func (m Memory) String() string {
	return strconv.Itoa(m.Int())
}

func (m Memory) Int() int {
	return int(m)
}

type Docker struct {
	SSH                       *ssh.SSH
	Name                      remote.ServiceName
	BuildDockerServiceContent func(ctx context.Context) (*DockerServiceContent, error)
}

func (d *Docker) Children() []world.Configuration {
	return []world.Configuration{
		&DockerEngine{
			SSH: d.SSH,
		},
		&Service{
			SSH:  d.SSH,
			Name: d.Name,
			Content: content.Func(func(ctx context.Context) ([]byte, error) {
				content, err := d.BuildDockerServiceContent(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "get DockerServiceContent failed")
				}
				return content.Content(ctx)
			}),
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
		d.Name,
	)
}

type Port struct {
	HostPort   network.Port
	DockerPort network.Port
}

func (v Port) String() string {
	return fmt.Sprintf("%v:%v", v.HostPort, v.DockerPort)
}

type Volume struct {
	HostPath   string
	DockerPath string
	Opts       string
}

func (v Volume) String() string {
	buf := bytes.Buffer{}
	buf.WriteString(v.HostPath)
	buf.WriteString(":")
	buf.WriteString(v.DockerPath)
	if v.Opts != "" {
		buf.WriteString(":")
		buf.WriteString(v.Opts)
	}
	return buf.String()
}

type DockerServiceContent struct {
	After            []remote.ServiceName
	Args             []string
	Before           []remote.ServiceName
	Command          string
	EnvironmentFiles []string
	HostNet          bool
	HostPid          bool
	Image            docker.Image
	Memory           Memory
	Name             remote.ServiceName
	Ports            []Port
	Privileged       bool
	Requires         []remote.ServiceName
	TimeoutStartSec  string
	TimeoutStopSec   string
	Volumes          []Volume
	UID              int
	GID              int
}

func (d *DockerServiceContent) Content(ctx context.Context) ([]byte, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "[Unit]\n")
	fmt.Fprintf(b, "Description=%s\n", d.Name)
	for _, service := range d.Requires {
		fmt.Fprintf(b, "Requires=%s\n", service)
	}
	for _, service := range d.After {
		fmt.Fprintf(b, "After=%s\n", service)
	}
	for _, service := range d.Before {
		fmt.Fprintf(b, "Before=%s\n", service)
	}
	fmt.Fprintf(b, "\n")
	fmt.Fprintf(b, "[Service]\n")
	fmt.Fprintf(b, "EnvironmentFile=/etc/environment\n")
	fmt.Fprintf(b, "Restart=always\n")
	fmt.Fprintf(b, "RestartSec=20s\n")
	if d.TimeoutStartSec != "" {
		fmt.Fprintf(b, "TimeoutStartSec=%s\n", d.TimeoutStartSec)
	}
	if d.TimeoutStopSec != "" {
		fmt.Fprintf(b, "TimeoutStopSec=%s\n", d.TimeoutStopSec)
	}
	fmt.Fprintf(b, "ExecStartPre=-/usr/bin/docker kill %s\n", d.Name)
	fmt.Fprintf(b, "ExecStartPre=-/usr/bin/docker rm %s\n", d.Name)
	fmt.Fprintf(b, "ExecStart=/usr/bin/docker run \\\n")
	fmt.Fprintf(b, "--memory-swap=0 \\\n")
	fmt.Fprintf(b, "--memory-swappiness=0 \\\n")
	fmt.Fprintf(b, "--memory=%dm \\\n", d.Memory)
	if d.Privileged {
		fmt.Fprintf(b, "--privileged=true \\\n")
	}
	if d.HostNet {
		fmt.Fprintf(b, "--net=host \\\n")
	}
	if d.HostPid {
		fmt.Fprintf(b, "--pid=host \\\n")
	}
	for _, file := range d.EnvironmentFiles {
		fmt.Fprintf(b, "--env-file %s \\\n", file)
	}
	for _, port := range d.Ports {
		dockerPort, err := port.DockerPort.Port(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get docker port failed")
		}
		hostPort, err := port.HostPort.Port(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get host port failed")
		}
		fmt.Fprintf(b, "-p %d:%d \\\n", hostPort, dockerPort)
	}
	for _, volume := range d.Volumes {
		fmt.Fprintf(b, "--volume=%s \\\n", volume.String())
	}
	if d.UID > 0 && d.GID > 0 {
		fmt.Fprintf(b, "--user %d:%d \\\n", d.UID, d.GID)
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
	return b.Bytes(), nil
}
