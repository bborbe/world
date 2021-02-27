// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service_test

import (
	"context"
	"github.com/bborbe/world/pkg/network"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bborbe/world/configuration/service"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/remote"
)

var _ = Describe("Docker", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
	})
	It("generate systemd service", func() {
		service := service.DockerServiceContent{
			Name:   "etcd",
			Memory: 1024,
			Ports: []service.Port{
				{
					HostPort:   network.PortStatic(2379),
					DockerPort: network.PortStatic(2379),
				},
				{
					HostPort:   network.PortStatic(2380),
					DockerPort: network.PortStatic(2380),
				},
			},
			Volumes: []service.Volume{
				{
					HostPath:   "/var/lib/etcd",
					DockerPath: "/var/lib/etcd",
					Opts:       "",
				},
			},
			Image: docker.Image{
				Repository: "quay.io/coreos/etcd",
				Tag:        "v3.3.1",
			},
			Command: "/usr/local/bin/etcd",
			Args: []string{
				"--advertise-client-urls http://172.16.22.10:2379",
				"--data-dir=/var/lib/etcd",
				"--initial-advertise-peer-urls http://172.16.22.10:2380",
				"--initial-cluster kubernetes=http://172.16.22.10:2380",
				"--initial-cluster-state new",
				"--initial-cluster-token cluster-fire",
				"--listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001",
				"--listen-peer-urls http://0.0.0.0:2380",
				"--name kubernetes",
			},
			Requires: []remote.ServiceName{
				"requires.service",
			},
			After: []remote.ServiceName{
				"after.service",
			},
			Before: []remote.ServiceName{
				"before.service",
			},
			TimeoutStartSec: "30s",
			TimeoutStopSec:  "10s",
		}
		bytes, err := service.Content(ctx)
		Expect(err).To(BeNil())
		Expect(string(bytes)).To(Equal(`[Unit]
Description=etcd
Requires=requires.service
After=after.service
Before=before.service

[Service]
EnvironmentFile=/etc/environment
Restart=always
RestartSec=20s
TimeoutStartSec=30s
TimeoutStopSec=10s
ExecStartPre=-/usr/bin/docker kill etcd
ExecStartPre=-/usr/bin/docker rm etcd
ExecStart=/usr/bin/docker run \
--memory-swap=0 \
--memory-swappiness=0 \
--memory=1024m \
-p 2379:2379 \
-p 2380:2380 \
--volume=/var/lib/etcd:/var/lib/etcd \
--name etcd \
quay.io/coreos/etcd:v3.3.1 \
/usr/local/bin/etcd \
--advertise-client-urls http://172.16.22.10:2379 \
--data-dir=/var/lib/etcd \
--initial-advertise-peer-urls http://172.16.22.10:2380 \
--initial-cluster kubernetes=http://172.16.22.10:2380 \
--initial-cluster-state new \
--initial-cluster-token cluster-fire \
--listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
--listen-peer-urls http://0.0.0.0:2380 \
--name kubernetes

ExecStop=/usr/bin/docker stop etcd

[Install]
WantedBy=multi-user.target
`))
	})
})
