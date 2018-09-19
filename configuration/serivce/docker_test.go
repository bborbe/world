package service_test

import (
	"github.com/bborbe/world/configuration/serivce"
	"github.com/bborbe/world/pkg/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Docker", func() {
	It("generate systemd service", func() {
		service := service.Docker{
			Name:    "etcd",
			Memory:  1024,
			Ports:   []int{2379, 2380},
			Volumes: []string{"/var/lib/etcd:/var/lib/etcd"},
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
		}
		bytes, err := service.SystemdServiceContent().Content()
		Expect(err).To(BeNil())
		Expect(string(bytes)).To(Equal(`[Unit]
Description=etcd
Requires=docker.service
After=docker.service

[Service]
Restart=always
RestartSec=20s
EnvironmentFile=/etc/environment
TimeoutStartSec=0
ExecStartPre=-/usr/bin/docker kill etcd
ExecStartPre=-/usr/bin/docker rm etcd
ExecStart=/usr/bin/docker run \
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
