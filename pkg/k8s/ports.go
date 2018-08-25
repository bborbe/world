package k8s

import (
	"context"

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

type PortProtocol string

type PortTarget string

type Port struct {
	Name       PortName     `yaml:"name,omitempty"`
	Port       PortNumber   `yaml:"port"`
	Protocol   PortProtocol `yaml:"protocol,omitempty"`
	TargetPort PortTarget   `yaml:"targetPort,omitempty"`
}
