// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Host string

func (h Host) String() string {
	return string(h)
}

type KeyPath string

func (k KeyPath) String() string {
	return string(k)
}

type Entry struct {
	Host Host
	IP   IP
}

type IP interface {
	IP(ctx context.Context) (net.IP, error)
	Validate(ctx context.Context) error
}

type IPStatic string

func (i IPStatic) IP(ctx context.Context) (net.IP, error) {
	return net.ParseIP(i.String()), nil
}

func (i IPStatic) String() string {
	return string(i)
}

func (i IPStatic) Validate(ctx context.Context) error {
	ip, err := i.IP(ctx)
	if err != nil {
		return err
	}
	if !ip.To4().Equal(ip) {
		return fmt.Errorf("ip '%s' is not a ipv4 addr", i.String())
	}
	return nil
}

type Server struct {
	Host    Host
	KeyPath KeyPath
	List    []Entry
}

func (s *Server) Apply(ctx context.Context) error {
	for _, entry := range s.List {
		ip, err := entry.IP.IP(ctx)
		if err != nil {
			return errors.Wrap(err, "get ip failed")
		}
		b := &bytes.Buffer{}
		fmt.Fprintf(b, "server %s\n", s.Host)
		fmt.Fprintf(b, "update delete %s 60 A\n", entry.Host.String())
		fmt.Fprintf(b, "update add %s 60 A %s\n", entry.Host.String(), ip.String())
		fmt.Fprintf(b, "send\n")
		glog.V(1).Infof("set dns for %s to %s", entry.Host.String(), ip.String())

		cmd := exec.CommandContext(ctx, "nsupdate", "-k", s.KeyPath.String())
		cmd.Stdin = b
		if glog.V(4) {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrap(err, "update dns failed")
		}
	}

	return nil
}

func (s *Server) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}

func (s *Server) Validate(ctx context.Context) error {
	for _, entry := range s.List {
		if entry.Host == "" {
			return errors.New("Host missing")
		}
		if err := entry.IP.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
