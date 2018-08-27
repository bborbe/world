package k8s

import (
	"context"

	"strconv"

	"github.com/pkg/errors"
)

type PortName string

func (p PortName) Validate(ctx context.Context) error {
	if p == "" {
		return errors.New("portName empty")
	}
	return nil
}

type PortNumber int

func (a PortNumber) Validate(ctx context.Context) error {
	if a == 0 {
		return errors.New("PortNumber missing")
	}
	return nil
}

func (c PortNumber) Int() int {
	return int(c)
}

func (c PortNumber) String() string {
	return strconv.Itoa(c.Int())
}

type PortProtocol string

func (p PortProtocol) Validate(ctx context.Context) error {
	if p != "TCP" && p != "UDP" {
		return errors.New("Protocol missing")
	}
	return nil
}

type PortTarget string

type Port struct {
	Name       PortName     `yaml:"name,omitempty"`
	Port       PortNumber   `yaml:"port"`
	Protocol   PortProtocol `yaml:"protocol,omitempty"`
	TargetPort PortTarget   `yaml:"targetPort,omitempty"`
}
