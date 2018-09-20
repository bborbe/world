package remote

import (
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/pkg/errors"
)

type Iptables struct {
	SSH  ssh.SSH
	Port k8s.PortNumber
}

func (f *Iptables) Satisfied(ctx context.Context) (bool, error) {
	return f.SSH.RunCommand(ctx, fmt.Sprintf("iptables -C INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", f.Port)) == nil, nil
}

func (f *Iptables) Apply(ctx context.Context) error {
	return errors.Wrap(f.SSH.RunCommand(ctx, fmt.Sprintf("iptables -A INPUT -p tcp -m state --state NEW -m tcp --dport %d -j ACCEPT", f.Port)), "iptables failed")
}

func (f *Iptables) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.SSH,
		f.Port,
	)
}
