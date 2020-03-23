// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hetzner

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/golang/glog"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type ServerType string

func (a ServerType) String() string {
	return string(a)
}

func (a ServerType) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("ServerType missing")
	}
	switch a.String() {
	case "cx11":
	case "cx21":
	case "cx31":
	case "cx41":
	case "cx51":
	default:
		return errors.New("ServerType invalid")
	}
	return nil
}

type ApiKey string

func (a ApiKey) String() string {
	return string(a)
}

func (a ApiKey) Client() *hcloud.Client {
	return hcloud.NewClient(hcloud.WithToken(a.String()))
}

func (a ApiKey) Validate(ctx context.Context) error {
	if a == "" {
		return errors.New("ApiKey missing")
	}
	return nil
}

type Server struct {
	ApiKey        deployer.SecretValue
	Name          k8s.Context
	PublicKeyPath ssh.PublicKeyPath
	User          ssh.User
	ServerType    ServerType
}

func (s *Server) Validate(ctx context.Context) error {
	return validation.Validate(ctx,
		s.ApiKey,
		s.Name,
		s.User,
		s.PublicKeyPath,
		s.ServerType,
	)
}

func (s *Server) Satisfied(ctx context.Context) (bool, error) {
	client, err := s.client(ctx)
	if err != nil {
		return false, err
	}
	server, _, err := client.Server.GetByName(ctx, s.Name.String())
	if err != nil {
		return false, errors.Wrapf(err, "get server %s failed", s.Name.String())
	}
	if server == nil {
		glog.V(1).Infof("server %s not found", s.Name.String())
		return false, nil
	}
	glog.V(2).Infof("server %s found", s.Name.String())
	return true, nil
}

func (s *Server) Apply(ctx context.Context) error {
	client, err := s.client(ctx)
	if err != nil {
		return err
	}
	datacenter, _, err := client.Datacenter.GetByID(ctx, 2)
	if err != nil {
		return errors.Wrap(err, "get datacenter 2 failed")
	}

	serverType, _, err := client.ServerType.GetByName(ctx, s.ServerType.String())
	if err != nil {
		return errors.Wrap(err, "get server type cx11 failed")
	}

	image, _, err := client.Image.GetByName(ctx, "ubuntu-18.04")
	if err != nil {
		return errors.Wrap(err, "get image ubuntu-18.04 failed")
	}

	sshKey, _, err := client.SSHKey.GetByName(ctx, "bborbe")
	if err != nil {
		return errors.Wrap(err, "get sshkey bborbe failed")
	}

	userdata, err := s.userdata()
	if err != nil {
		return err
	}

	start := true
	glog.V(1).Infof("create server %s on hetzner cloud", s.Name.String())
	_, _, err = client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name:             s.Name.String(),
		ServerType:       serverType,
		Image:            image,
		SSHKeys:          []*hcloud.SSHKey{sshKey},
		Location:         datacenter.Location,
		UserData:         userdata,
		StartAfterCreate: &start,
		Labels:           map[string]string{},
	})
	return errors.Wrapf(err, "create hetzner server %s failed", s.Name.String())
}

func (s *Server) client(ctx context.Context) (*hcloud.Client, error) {
	bytes, err := s.ApiKey.Value(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get secret failed")
	}
	return ApiKey(bytes).Client(), nil
}

type UserData struct {
	Users []User `yaml:"users"`
}

type User struct {
	Name              string   `yaml:"name"`
	Sudo              string   `yaml:"sudo"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Shell             string   `yaml:"shell"`
}

func (s *Server) userdata() (string, error) {
	key, err := ioutil.ReadFile(s.PublicKeyPath.String())
	if err != nil {
		return "", errors.Wrap(err, "read ssh key failed")
	}
	b := bytes.NewBufferString("#cloud-config\n")
	err = yaml.NewEncoder(b).Encode(UserData{
		Users: []User{
			{
				Name: s.User.String(),
				Sudo: "ALL=(ALL) NOPASSWD:ALL",
				SshAuthorizedKeys: []string{
					string(key),
				},
				Shell: "/bin/bash",
			},
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "encode yaml failed")
	}
	return b.String(), nil
}
