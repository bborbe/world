package server

import (
	"context"
	"fmt"
	"net"

	"github.com/bborbe/world/pkg/remote"

	"github.com/bborbe/world/pkg/configuration"
	"github.com/bborbe/world/pkg/dns"

	"github.com/bborbe/world/pkg/k8s"

	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Nuke struct {
	Context   k8s.Context
	ClusterIP k8s.ClusterIP
}

func (n *Nuke) Children() []world.Configuration {
	ssh := ssh.SSH{
		Host:           ssh.Host(fmt.Sprintf("%s:22", n.ClusterIP)),
		User:           "bborbe",
		PrivateKeyPath: "/Users/bborbe/.ssh/id_rsa",
	}
	return []world.Configuration{
		configuration.New().WithApplier(
			&dns.Server{
				Host:    "ns.rocketsource.de",
				KeyPath: "/Users/bborbe/.dns/home.benjamin-borbe.de.key",
				List: []dns.Entry{
					{
						Host: "backup.nuke.hm.benjamin-borbe.de",
						IP:   net.ParseIP(n.ClusterIP.String()),
					},
				},
			},
		),
		configuration.New().WithApplier(&remote.Iptables{
			SSH:  ssh,
			Port: 80,
		}),
		&service.Kubernetes{
			SSH:       ssh,
			Context:   n.Context,
			ClusterIP: n.ClusterIP,
		},
	}
}

func (n *Nuke) Applier() (world.Applier, error) {
	return nil, nil
}

func (n *Nuke) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		n.Context,
		n.ClusterIP,
	)
}
