// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

type PortName string

func (p PortName) Validate(ctx context.Context) error {
	if p == "" {
		return errors.New("portName empty")
	}
	portNameRe := regexp.MustCompile("^[a-z0-9-]+$")
	if !portNameRe.MatchString(p.String()) {
		return fmt.Errorf("portName %s invalid", p.String())
	}
	return nil
}

func (p PortName) String() string {
	return string(p)
}

type PortNumber int

func (p PortNumber) Validate(ctx context.Context) error {
	if p <= 0 || p >= 65535 {
		return errors.New("PortNumber invalid")
	}
	return nil
}

func (p PortNumber) Int() int {
	return int(p)
}

func (p PortNumber) String() string {
	return strconv.Itoa(p.Int())
}

type PortProtocol string

func (p PortProtocol) String() string {
	return string(p)
}

func (p PortProtocol) Validate(ctx context.Context) error {
	if p != "TCP" && p != "UDP" {
		return errors.New("Protocol missing")
	}
	return nil
}

type PortTarget string

type ServicePort struct {
	Name       PortName     `yaml:"name,omitempty"`
	Port       PortNumber   `yaml:"port"`
	Protocol   PortProtocol `yaml:"protocol,omitempty"`
	TargetPort PortTarget   `yaml:"targetPort,omitempty"`
}
