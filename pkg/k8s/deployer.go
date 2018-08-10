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

type DataProvider interface {
	Data() (interface{}, error)
}

type Deployer struct {
	Context world.Context
	Data    DataProvider
}

func (d *Deployer) Apply(ctx context.Context) error {
	data, err := d.Data.Data()
	if err != nil {
		return errors.Wrap(err, "get data failed")
	}
	glog.V(2).Infof("deploy %s to %s ...", data, d.Context)
	buf := &bytes.Buffer{}
	if err := yaml.NewEncoder(buf).Encode(data); err != nil {
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
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "deploy %T to %s failed", data, d.Context)

	}
	glog.V(1).Infof("deploy %s to %s finished", data, d.Context)
	return nil
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
