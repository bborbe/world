package k8s

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"fmt"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/uploader"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context   world.Context
	Namespace world.Namespace
	Domains   []world.Domain
	Args      []world.Arg
	Port      world.Port
	Env       world.Env
	Uploader  world.Uploader
}

func (d *Deployer) Deploy(ctx context.Context) error {
	glog.V(2).Infof("deploy %s to %s ...", d.Namespace, d.Context)

	if err := uploader.UploadIfNeeded(ctx, d.Uploader); err != nil {
		return err
	}

	if err := d.apply(ctx, `{{ $out := . }}
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
`, struct {
		Name string
	}{
		Name: d.Namespace.String(),
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := d.apply(ctx, `{{ $out := . }}
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
		Name world.Namespace
	}{
		Name: d.Namespace,
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := d.apply(ctx, `{{ $out := . }}
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
		Name    world.Namespace
		Domains []world.Domain
	}{
		Name:    d.Namespace,
		Domains: d.Domains,
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	if err := d.apply(ctx, `{{ $out := . }}
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
        image: {{ .Image.String }}
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
		Name  world.Namespace
		Image world.Image
		Args  []world.Arg
		Port  world.Port
		Env   world.Env
	}{
		Name:  d.Namespace,
		Image: d.Uploader.GetBuilder().GetImage(),
		Args:  d.Args,
		Env:   d.Env,
		Port:  d.Port,
	}); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}
	glog.V(2).Infof("deploy %s to %s finished", d.Namespace, d.Context)
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

func (d *Deployer) Validate(ctx context.Context) error {
	if d.Uploader == nil {
		return fmt.Errorf("%s has no builder", d.Namespace)
	}
	if err := d.Uploader.Validate(ctx); err != nil {
		return err
	}
	if d.Context == "" {
		return errors.New("context missing")
	}
	if d.Namespace == "" {
		return errors.New("namespace missing")
	}
	if len(d.Domains) == 0 {
		return errors.New("domains empty")
	}
	if d.Port <= 0 || d.Port > 65535 {
		return errors.New("port missing")
	}
	return nil
}

func (d *Deployer) GetUploader() world.Uploader {
	return d.Uploader
}

func (d *Deployer) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}
