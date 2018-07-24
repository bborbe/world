package k8s

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"text/template"

	"fmt"

	"io"

	"github.com/bborbe/world"
	"github.com/bborbe/world/pkg/uploader"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Deployer struct {
	Context        world.Context
	Namespace      world.Namespace
	Domains        []world.Domain
	Args           []world.Arg
	Port           world.Port
	HostPort       world.HostPort
	Env            world.Env
	Uploader       world.Uploader
	CpuLimit       world.CpuLimit
	MemoryLimit    world.MemoryLimit
	CpuRequest     world.CpuRequest
	MemoryRequest  world.MemoryRequest
	LivenessProbe  world.LivenessProbe
	ReadinessProbe world.ReadinessProbe
	Mounts         []world.Mount
}

func (d *Deployer) Deploy(ctx context.Context) error {
	glog.V(2).Infof("deploy %s to %s ...", d.Namespace, d.Context)

	if err := uploader.UploadIfNeeded(ctx, d.Uploader); err != nil {
		return err
	}

	namespace, err := d.generateNamespace()
	if err != nil {
		return errors.Wrap(err, "generate namespace failed")
	}
	if err := d.apply(ctx, namespace); err != nil {
		return errors.Wrap(err, "apply namespace failed")
	}

	service, err := d.generateService()
	if err != nil {
		return errors.Wrap(err, "generate service failed")
	}
	if err := d.apply(ctx, service); err != nil {
		return errors.Wrap(err, "apply service failed")
	}

	if len(d.Domains) > 0 {
		ingress, err := d.generateIngress()
		if err != nil {
			return errors.Wrap(err, "generate ingress failed")
		}
		if err := d.apply(ctx, ingress); err != nil {
			return errors.Wrap(err, "apply deployment failed")
		}
	}

	deployment, err := d.generateDeployment()
	if err != nil {
		return errors.Wrap(err, "generate deployment failed")
	}
	if err := d.apply(ctx, deployment); err != nil {
		return errors.Wrap(err, "apply deployment failed")
	}

	glog.V(2).Infof("deploy %s to %s finished", d.Namespace, d.Context)
	return nil
}

func (d *Deployer) generateNamespace() (io.Reader, error) {
	return generateTemplate(`{{ $out := . }}
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app: {{ .Name }}
  name: {{ .Name }}
`, struct {
		Name world.Namespace
	}{
		Name: d.Namespace,
	})
}

func (d *Deployer) generateService() (io.Reader, error) {
	return generateTemplate(`{{ $out := . }}
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
    port: {{ .Port }}
    protocol: TCP
    targetPort: http
  selector:
    app: {{ .Name }}
`, struct {
		Name world.Namespace
		Port world.Port
	}{
		Name: d.Namespace,
		Port: d.Port,
	})
}

func (d *Deployer) generateIngress() (io.Reader, error) {
	return generateTemplate(`{{ $out := . }}
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
	})
}

func (d *Deployer) generateDeployment() (io.Reader, error) {
	return generateTemplate(`{{ $out := . }}
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
{{ if gt .HostPort 0 }} 
          hostPort: {{ .HostPort }}
{{ end }}
          name: http
          protocol: TCP
{{ if .LivenessProbe }} 
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /
            port: {{ $out.Port }}
            scheme: HTTP
          initialDelaySeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
{{ end }}
{{ if .ReadinessProbe }} 
        readinessProbe:
          httpGet:
            path: /
            port: {{ $out.Port }}
            scheme: HTTP
          initialDelaySeconds: 10
          timeoutSeconds: 5
{{ end }}
        resources:
          limits:
            cpu: {{ .CpuLimit }}
            memory: {{ .MemoryLimit }}
          requests:
            cpu: {{ .CpuRequest }}
            memory: {{ .MemoryRequest }}
{{ if gt (len .Mounts) 0 }} 
        volumeMounts:
{{ range $mount := .Mounts }}
        - mountPath: {{ $mount.Target }}
          name: {{ $mount.Name }}
          readOnly: {{ $mount.ReadOnly }}
{{ end }}
      volumes:
{{ range $mount := .Mounts }}
      - name: {{ $mount.Name }}
        nfs:
          path: '{{ $mount.NfsPath }}'
          server: '{{ $mount.NfsServer }}'
{{ end }}
{{ end }}
`, struct {
		Name           world.Namespace
		Image          world.Image
		Args           []world.Arg
		Port           world.Port
		HostPort       world.HostPort
		Env            world.Env
		CpuLimit       world.CpuLimit
		MemoryLimit    world.MemoryLimit
		CpuRequest     world.CpuRequest
		MemoryRequest  world.MemoryRequest
		LivenessProbe  world.LivenessProbe
		ReadinessProbe world.ReadinessProbe
		Mounts         []world.Mount
	}{
		Name:           d.Namespace,
		Image:          d.Uploader.GetBuilder().GetImage(),
		Args:           d.Args,
		Env:            d.Env,
		Port:           d.Port,
		HostPort:       d.HostPort,
		CpuLimit:       d.CpuLimit,
		MemoryLimit:    d.MemoryLimit,
		CpuRequest:     d.CpuRequest,
		MemoryRequest:  d.MemoryRequest,
		LivenessProbe:  d.LivenessProbe,
		ReadinessProbe: d.ReadinessProbe,
		Mounts:         d.Mounts,
	})
}

func generateTemplate(manifest string, data interface{}) (io.Reader, error) {
	tmpl, err := template.New("template").Parse(manifest)
	if err != nil {
		return nil, errors.Wrap(err, "parse k8s manifest failed")
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, errors.Wrap(err, "fill k8sfile template failed")
	}
	return buf, nil
}

func (d *Deployer) apply(ctx context.Context, buf io.Reader) error {
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
	for _, domain := range d.Domains {
		if domain == "" {
			return errors.New("domain empty")
		}
	}
	if d.Port <= 0 || d.Port > 65535 {
		return errors.New("port missing")
	}
	if d.CpuLimit == "" {
		return errors.New("cpu limit missing")
	}
	if d.MemoryLimit == "" {
		return errors.New("memory limit missing")
	}
	if d.CpuRequest == "" {
		return errors.New("cpu request missing")
	}
	if d.MemoryRequest == "" {
		return errors.New("memory request missing")
	}
	for _, mount := range d.Mounts {
		if err := mount.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deployer) GetUploader() world.Uploader {
	return d.Uploader
}

func (d *Deployer) Satisfied(ctx context.Context) (bool, error) {
	return false, nil
}
