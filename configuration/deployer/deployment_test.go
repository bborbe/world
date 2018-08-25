package deployer

import (
	"bytes"

	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DeploymentDeployer", func() {
	format.TruncatedDiff = false
	var deploymentDeployer *DeploymentDeployer
	BeforeEach(func() {
		deploymentDeployer = &DeploymentDeployer{
			Namespace: "banana",
			Name:      "banana",
			Containers: []DeploymentDeployerContainer{
				{
					Name: "banana",
					Image: docker.Image{
						Repository: "bborbe/test",
						Tag:        "latest",
					},
					Ports: []Port{
						{
							Name:     "root",
							Port:     1337,
							Protocol: "TCP",
						},
					},
					Args: []k8s.Arg{"-v=4"},
					Env: []k8s.Env{
						{
							Name:  "a",
							Value: "b",
						},
					},
					Resources: k8s.Resources{
						Limits: k8s.ContainerResource{
							Cpu:    "250m",
							Memory: "25Mi",
						},
						Requests: k8s.ContainerResource{
							Cpu:    "10m",
							Memory: "10Mi",
						},
					},
					Mounts: []k8s.ContainerMount{
						{
							Name:     "data",
							Path:     "/usr/share/nginx/html",
							ReadOnly: true,
						},
					},
				},
			},
			Volumes: []k8s.PodVolume{
				{
					Name: "data",
					Nfs: k8s.PodVolumeNfs{
						Path:   "/data/download",
						Server: "127.0.0.1",
					},
				},
			},
		}
	})
	Context("with hostPort", func() {
		BeforeEach(func() {
			deploymentDeployer.Containers[0].Ports[0].HostPort = 123
		})
		It("generateDeployment contains hostport", func() {
			b := &bytes.Buffer{}
			err := yaml.NewEncoder(b).Encode(deploymentDeployer.deployment())
			Expect(err).NotTo(HaveOccurred())
			Expect(gbytes.BufferWithBytes(b.Bytes())).To(gbytes.Say("hostPort: 123"))
		})
	})
	It("deployment", func() {
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(deploymentDeployer.deployment())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
spec:
  replicas: 1
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      app: banana
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app: banana
    spec:
      containers:
      - name: banana
        image: bborbe/test:latest
        args:
        - -v=4
        env:
        - name: a
          value: b
        ports:
        - containerPort: 1337
          name: root
          protocol: TCP
        resources:
          limits:
            cpu: 250m
            memory: 25Mi
          requests:
            cpu: 10m
            memory: 10Mi
        volumeMounts:
        - mountPath: /usr/share/nginx/html
          name: data
          readOnly: true
      volumes:
      - name: data
        nfs:
          path: /data/download
          server: 127.0.0.1
`))
	})
})
