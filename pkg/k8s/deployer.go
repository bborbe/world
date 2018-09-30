package k8s

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/go-yaml/yaml"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context Context
	Data    interface {
		String() string
	}
}

func (d *Deployer) Apply(ctx context.Context) error {
	glog.V(3).Infof("deploy %s to %s ...", d.Data, d.Context)
	buf := &bytes.Buffer{}
	if err := yaml.NewEncoder(buf).Encode(d.Data); err != nil {
		return err
	}
	if glog.V(4) {
		glog.Infof("yaml: %s", buf.String())
	}
	glog.V(1).Infof("kubectl apply %s", d.Data.String())
	cmd := exec.CommandContext(ctx, "kubectl", "--context", d.Context.String(), "apply", "-f", "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "deploy %T to %s failed", d.Data, d.Context)

	}
	glog.V(3).Infof("deploy %s to %s finished", d.Data, d.Context)
	return nil
}
