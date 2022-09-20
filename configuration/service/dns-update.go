// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bborbe/world/pkg/apt"
	"github.com/bborbe/world/pkg/content"
	"github.com/bborbe/world/pkg/deployer"
	"github.com/bborbe/world/pkg/file"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type DnsUpdate struct {
	SSH        *ssh.SSH
	DnsKey     deployer.SecretValue
	DnsPrivate deployer.SecretValue
	DnsZone    string
	DnsName    string
}

func (d *DnsUpdate) Children(ctx context.Context) (world.Configurations, error) {
	return world.Configurations{
		&remote.File{
			SSH:     d.SSH,
			Path:    file.Path("/usr/local/bin/dns-update.sh"),
			Content: d.dnsUpdateContent(),
			User:    "root",
			Group:   "root",
			Perm:    0500,
		},

		world.NewConfiguraionBuilder().WithApplier(&remote.Directory{
			SSH:  d.SSH,
			Path: file.Path("/etc/dns-update/keys"),
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   d.SSH,
			Path:  file.Path("/etc/dns-update/keys"),
			User:  "root",
			Group: "root",
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chmod{
			SSH:  d.SSH,
			Path: file.Path("/etc/dns-update/keys"),
			Perm: 0700,
		}),

		world.NewConfiguraionBuilder().WithApplier(&remote.Directory{
			SSH:  d.SSH,
			Path: file.Path("/var/log/dns-update"),
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chown{
			SSH:   d.SSH,
			Path:  file.Path("/var/log/dns-update"),
			User:  "root",
			Group: "root",
		}),
		world.NewConfiguraionBuilder().WithApplier(&remote.Chmod{
			SSH:  d.SSH,
			Path: file.Path("/var/log/dns-update"),
			Perm: 0700,
		}),

		&remote.File{
			SSH:     d.SSH,
			Path:    file.Path(fmt.Sprintf("/etc/dns-update/keys/%s.%s.private", d.DnsName, d.DnsZone)),
			Content: content.Func(d.DnsPrivate.Value),
			User:    "root",
			Group:   "root",
			Perm:    0400,
		},
		&remote.File{
			SSH:     d.SSH,
			Path:    file.Path(fmt.Sprintf("/etc/dns-update/keys/%s.%s.key", d.DnsName, d.DnsZone)),
			Content: content.Func(d.DnsKey.Value),
			User:    "root",
			Group:   "root",
			Perm:    0400,
		},
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "curl",
		}),
		world.NewConfiguraionBuilder().WithApplier(&apt.Install{
			SSH:     d.SSH,
			Package: "dnsutils",
		}),
		&Cron{
			SSH:        d.SSH,
			Name:       d.cronName(),
			Expression: d.cronExpression(),
			Schedule:   "* * * * *",
		},
	}, nil
}

func (d *DnsUpdate) cronName() CronName {
	return BuildCronName("dns-update", d.DnsName, d.DnsZone)
}

func (d *DnsUpdate) cronExpression() content.HasContent {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "/usr/local/bin/dns-update.sh ")
	fmt.Fprintf(buf, "ns.rocketsource.de ")
	fmt.Fprintf(buf, "/etc/dns-update/keys/%s.%s ", d.DnsName, d.DnsZone)
	fmt.Fprintf(buf, "%s ", d.DnsZone)
	fmt.Fprintf(buf, "%s ", d.DnsName)
	fmt.Fprintf(buf, "https://ip.benjamin-borbe.de ")
	fmt.Fprintf(buf, ">> /var/log/dns-update/%s.%s.log", d.DnsName, d.DnsZone)
	return content.Static(buf.String())
}

func (d *DnsUpdate) dnsUpdateContent() content.HasContent {
	return content.Static(`#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

server="$1"
key="$2"
zone="$3"
node="$4"
ip=$(curl -s $5)

echo "update ${zone}.${node} with $ip started"

ttl='60'
class='A'
tmpfile=$(mktemp)
cat >$tmpfile <<END
server $server
update delete ${node}.${zone} $ttl $class
update add ${node}.${zone} $ttl $class $ip
send
END
nsupdate -k $key -v $tmpfile
rm -f $tmpfile

echo "update ${zone}.${node} with $ip finished"
`)
}

func (d *DnsUpdate) Applier() (world.Applier, error) {
	return nil, nil
}

func (d *DnsUpdate) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		d.SSH,
	)
}
