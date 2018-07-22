package k8s

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"github.com/bborbe/world"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context world.Context
	Name    world.Name
	Image   world.Image
	Domains []world.Domain
	Args    []world.Arg
	Port    world.Port
	Env     world.Env
}

func (b *Deployer) Deploy(ctx context.Context) error {
	glog.V(2).Infof("deploy %s to %s ...", b.Name, b.Context)

	if err := b.apply(ctx, `{{ $out := . }}
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
`, struct {
		Name string
	}{
		Name: b.Name.String(),
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := b.apply(ctx, `{{ $out := . }}
apiVersion: v1
kind: Service
metadata:
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
  namespace: {{ .Name }}
spec:
  ports:
  - name: web
    port: 80
    protocol: TCP
    targetPort: http
  selector:
    app: {{ .Name }}
`, struct {
		Name string
	}{
		Name: b.Name.String(),
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := b.apply(ctx, `{{ $out := . }}
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.priority: "10000"
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
  namespace: {{ .Name }}
spec:
  rules:
{{ range $domain := .Domains }}
  - host: {{ $domain }}
    http:
      paths:
      - backend:
          serviceName: {{ $out.Name }}
          servicePort: web
        path: /
{{ end }}
`, struct {
		Name    world.Name
		Domains []world.Domain
	}{
		Name:    b.Name,
		Domains: b.Domains,
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := b.apply(ctx, `{{ $out := . }}
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
  namespace: {{ .Name }}
spec:
  replicas: 1
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      app: {{ .Name }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: {{ .Name }}
    spec:
      containers:
      - args:
{{ range $arg := .Args }}
        - {{ $arg }}
{{ end }}
        env:
{{ range $key, $value := .Env }}
        - name: {{ $key }}
          value: "{{ $value }}"
{{ end }}
        image: {{ .Image }}
        name: {{ .Name }}
        ports:
        - containerPort: {{ .Port }}
          name: http
          protocol: TCP
        resources:
          limits:
            cpu: 100m
            memory: 50Mi
          requests:
            cpu: 10m
            memory: 10Mi
`, struct {
		Name  string
		Image string
		Args  []world.Arg
		Port  world.Port
		Env   world.Env
	}{
		Name:  b.Name.String(),
		Image: b.Image.String(),
		Args:  b.Args,
		Env:   b.Env,
		Port:  b.Port,
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	return nil
}

func (d *Deployer) apply(ctx context.Context, manifest string, data interface{}) error {
	tmpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return errors.Wrap(err, "parse k8s manifest failed")
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return errors.Wrap(err, "fill k8sfile template failed")
	}
	cmd := exec.CommandContext(ctx, "kubectl", "--context", d.Context.String(), "apply", "-f", "-")
	cmd.Stdin = buf
	if glog.V(4) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return errors.Wrap(cmd.Run(), "deploy k8s image failed")
}
