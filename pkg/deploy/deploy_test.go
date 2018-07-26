package deploy

import (
	"bytes"
	"testing"

	"github.com/bborbe/world"
	"github.com/go-yaml/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Deployer", func() {
	format.TruncatedDiff = true
	var deployer *Deployer
	BeforeEach(func() {
		deployer = &Deployer{
			Image: world.Image{
				Registry:   "docker.io",
				Repository: "bborbe/test",
				Tag:        "latest",
			},
			Port:      1337,
			Namespace: "banana",
			Domains:   []world.Domain{"example.com"},
			Args:      []world.Arg{"-v=4"},
			Env: world.Env{
				"a": "b",
			},
			CpuLimit:      "250m",
			MemoryLimit:   "25Mi",
			CpuRequest:    "10m",
			MemoryRequest: "10Mi",
			Mounts: []world.Mount{
				{
					Name:      "data",
					Target:    "/usr/share/nginx/html",
					ReadOnly:  true,
					NfsPath:   "/data/download",
					NfsServer: "127.0.0.1",
				},
			},
		}
	})
	Context("with hostPort", func() {
		BeforeEach(func() {
			deployer.HostPort = 123
		})
		It("generateDeployment contains hostport", func() {
			b := &bytes.Buffer{}
			err := yaml.NewEncoder(b).Encode(deployer.deployment())
			Expect(err).NotTo(HaveOccurred())
			Expect(gbytes.BufferWithBytes(b.Bytes())).To(gbytes.Say("hostPort: 123"))
		})
	})
	It("namespace", func() {
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(deployer.namespace())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Namespace
metadata:
  name: banana
  labels:
    app: banana
`))
	})
	It("service", func() {
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(deployer.service())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: v1
kind: Service
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
spec:
  ports:
  - name: web
    port: 1337
    protocol: TCP
    targetPort: http
  selector:
    app: banana
`))
	})

	It("ingress", func() {
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(deployer.ingress())
		Expect(err).NotTo(HaveOccurred())
		Expect(b.String()).To(Equal(`apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: banana
  name: banana
  labels:
    app: banana
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.priority: "10000"
spec:
  rules:
  - host: example.com
    http:
      paths:
      - backend:
          serviceName: banana
          servicePort: web
        path: /
`))
	})

	It("deployment", func() {
		b := &bytes.Buffer{}
		err := yaml.NewEncoder(b).Encode(deployer.deployment())
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
      - args:
        - -v=4
        env:
        - name: a
          value: b
        image: docker.io/bborbe/test:latest
        name: banana
        ports:
        - containerPort: 1337
          name: http
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

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s Deploy Suite")
}
