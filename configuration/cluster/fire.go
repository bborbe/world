package cluster

import (
	"context"

	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Fire struct {
	Context     k8s.Context
	ClusterIP   dns.IP
	DisableRBAC bool
}

func (f *Fire) Children() []world.Configuration {
	ssh := ssh.SSH{
		Host: ssh.Host{
			IP:   f.ClusterIP,
			Port: 22,
		},
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "backup.fire.hm.benjamin-borbe.de",
						IP:   f.ClusterIP,
					},
				},
			},
		),
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  ssh,
			Port: 80,
		}),
		&service.Kubernetes{
			SSH:         ssh,
			Context:     f.Context,
			ClusterIP:   f.ClusterIP,
			DisableRBAC: f.DisableRBAC,
		},
	}
}

func (f *Fire) Applier() (world.Applier, error) {
	return nil, nil
}

func (f *Fire) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		f.Context,
		f.ClusterIP,
	)
}