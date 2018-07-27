package k8s

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/bborbe/world"
	"github.com/go-yaml/yaml"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context world.Context
	Data    interface{}
}

func (d *Deployer) Deploy(ctx context.Context) error {
	buf := &bytes.Buffer{}
	if err := yaml.NewEncoder(buf).Encode(d.Data); err != nil {
		return err
	}
	if glog.V(4) {
		glog.Infof("yaml: %s", buf.String())
	}
	cmd := exec.CommandContext(ctx, "kubectl", "--context", d.Context.String(), "apply", "-f", "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "deploy k8s image failed")
}

func (d *Deployer) Validate(ctx context.Context) error {
	if d.Context == "" {
		return fmt.Errorf("context missing")
	}
	if d.Data == nil {
		return fmt.Errorf("data missing")
	}
	return nil
}

func (d *Deployer) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}
