// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/dns"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/k8s"
	"github.com/bborbe/world/pkg/local"
	"github.com/bborbe/world/pkg/remote"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/cert/triple"
)

const caKey = "ca-key.pem"
const caCert = "ca-cert.pem"
const serverKey = "server-key.pem"
const serverCert = "server-cert.pem"
const clientKey = "client-key.pem"
const clientCert = "client-cert.pem"

type Kubelet struct {
	SSH         *ssh.SSH
	Version     docker.Tag
	Context     k8s.Context
	ClusterIP   dns.IP
	DisableRBAC bool
	DisableCNI  bool
	ResolvConf  string
}

func (k *Kubelet) Children() []world.Configuration {
	return []world.Configuration{
		world.NewConfiguraionBuilder().WithApplier(&remote.Iptables{
			SSH:  k.SSH,
			Port: 6443,
		}),
		&build.Hyperkube{
			Image: k.hyperkubeImage(),
		},
		&build.Podmaster{
			Image: k.podmasterImage(),
		},
		&build.Pause{
			Image: k.pauseImage(),
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes/ssl",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes/manifests",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/srv",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/srv/kubernetes",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/srv/kubernetes/manifests",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&Directory{
			SSH:   k.SSH,
			Path:  "/var/lib/kubelet",
			User:  "root",
			Group: "root",
			Perm:  0755,
		},
		&File{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes/ssl/ca.pem",
			User:  "root",
			Group: "root",
			Perm:  0644,
			Content: remote.ContentFunc(func(ctx context.Context) ([]byte, error) {
				return k.readPem(ctx, caCert)
			}),
		},
		&File{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes/ssl/node-key.pem",
			User:  "root",
			Group: "root",
			Perm:  0644,
			Content: remote.ContentFunc(func(ctx context.Context) ([]byte, error) {
				return k.readPem(ctx, serverKey)
			}),
		},
		&File{
			SSH:   k.SSH,
			Path:  "/etc/kubernetes/ssl/node.pem",
			User:  "root",
			Group: "root",
			Perm:  0644,
			Content: remote.ContentFunc(func(ctx context.Context) ([]byte, error) {
				return k.readPem(ctx, serverCert)
			}),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/manifests/kube-apiserver.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubeApiserverYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/manifests/kube-podmaster.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubePodmasterYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/kubeconfig.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.nodeKubeconfigYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/manifests/kube-proxy.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubeProxyYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/srv/kubernetes/manifests/kube-scheduler.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubeSchedulerYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/srv/kubernetes/manifests/kube-controller-manager.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubeControllerManagerYaml),
		},
		&Docker{
			SSH:  k.SSH,
			Name: "kubelet",
			BuildDockerServiceContent: func(ctx context.Context) (*DockerServiceContent, error) {
				ip, err := k.ClusterIP.IP(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "get ip failed")
				}

				args := []string{
					"kubelet",
					"--fail-swap-on=false",
					fmt.Sprintf("--pod-infra-container-image=%s", k.pauseImage().String()),
					"--containerized",
					"--register-node=true",
					"--allow-privileged=true",
					"--pod-manifest-path=/etc/kubernetes/manifests",
					fmt.Sprintf("--hostname-override=%s", ip.String()),
					"--cluster-dns=10.103.0.10",
					"--cluster-domain=cluster.local",
					"--kubeconfig=/etc/kubernetes/kubeconfig.yaml",
					"--node-labels=etcd=true,nfsd=true,worker=true,master=true",
					"--v=0",
				}
				if k.ResolvConf != "" {
					args = append(args, fmt.Sprintf("--resolv-conf=%s", k.ResolvConf))
				}
				if !k.DisableCNI {
					args = append(args, "--network-plugin=cni", "--cni-conf-dir=/etc/cni/net.d", "--cni-bin-dir=/opt/cni/bin")
				}
				volumes := []string{
					"/:/rootfs:ro",
					"/sys:/sys:ro",
					"/var/log/:/var/log:rw",
					"/var/lib/docker/:/var/lib/docker:rw",
					"/var/lib/kubelet/:/var/lib/kubelet:rw,rslave",
					"/run:/run:rw",
					"/var/run:/var/run:rw",
					"/etc/kubernetes:/etc/kubernetes",
					"/srv/kubernetes:/srv/kubernetes",
				}
				if !k.DisableCNI {
					volumes = append(volumes, "/etc/cni/net.d:/etc/cni/net.d", "/opt/cni/bin:/opt/cni/bin", "/var/lib/calico:/var/lib/calico")
				}

				return &DockerServiceContent{
					Name:       "kubelet",
					Memory:     2048,
					HostNet:    true,
					HostPid:    true,
					Privileged: true,
					Volumes:    volumes,
					Image:      k.hyperkubeImage(),
					Command:    "/hyperkube",
					Args:       args,
				}, nil
			},
		},
		world.NewConfiguraionBuilder().WithApplierBuildFunc(func(ctx context.Context) (world.Applier, error) {
			ip, err := k.ClusterIP.IP(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "get ip failed")
			}
			return &local.Command{
				Command: "kubectl",
				Args: []string{
					"config",
					"set-cluster",
					fmt.Sprintf("%s-cluster", k.Context),
					fmt.Sprintf("--server=https://%s:6443", ip.String()),
					fmt.Sprintf("--certificate-authority=/Users/bborbe/.kube/%s/%s", k.Context, caCert),
				},
			}, nil
		}),
		world.NewConfiguraionBuilder().WithApplier(&local.Command{
			Command: "kubectl",
			Args: []string{
				"config",
				"set-credentials",
				fmt.Sprintf("%s-admin", k.Context),
				fmt.Sprintf("--certificate-authority=/Users/bborbe/.kube/%s/%s", k.Context, caCert),
				fmt.Sprintf("--client-key=/Users/bborbe/.kube/%s/%s", k.Context, clientKey),
				fmt.Sprintf("--client-certificate=/Users/bborbe/.kube/%s/%s", k.Context, clientCert),
			},
		}),
		world.NewConfiguraionBuilder().WithApplier(&local.Command{
			Command: "kubectl",
			Args: []string{
				"config",
				"set-context",
				k.Context.String(),
				fmt.Sprintf("--cluster=%s-cluster", k.Context),
				fmt.Sprintf("--user=%s-admin", k.Context),
			},
		}),
	}
}

func (k *Kubelet) Applier() (world.Applier, error) {
	return nil, nil
}

func (k *Kubelet) Validate(ctx context.Context) error {
	return validation.Validate(
		ctx,
		k.SSH,
		k.Version,
		k.Context,
		k.ClusterIP,
	)
}

func (k *Kubelet) kubeApiserverYaml(ctx context.Context) ([]byte, error) {
	ip, err := k.ClusterIP.IP(ctx)
	if err != nil {
		return nil, err
	}
	return render(`  
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver
  namespace: kube-system
spec:
  hostNetwork: true
  containers:
  - name: kube-apiserver
    image: {{.Image}}
    command:
    - /hyperkube
    - apiserver
    - --bind-address=0.0.0.0
    - --etcd-servers=http://{{.ClusterIP}}:2379
    - --storage-backend=etcd3
    - --allow-privileged=true
    - --service-cluster-ip-range=10.103.0.0/16
    - --secure-port=6443
    - --advertise-address={{.ClusterIP}}
    - --admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota
    - --tls-cert-file=/etc/kubernetes/ssl/node.pem
    - --tls-private-key-file=/etc/kubernetes/ssl/node-key.pem
    - --client-ca-file=/etc/kubernetes/ssl/ca.pem
    - --service-account-key-file=/etc/kubernetes/ssl/node-key.pem
    - --runtime-config=extensions/v1beta1/networkpolicies=true,batch/v2alpha1=true
    - --anonymous-auth=false
{{- if .RBAC }} 
    - --authorization-mode=RBAC
{{- end }}
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        port: 8080
        path: /healthz
      initialDelaySeconds: 15
      timeoutSeconds: 15
    ports:
    - containerPort: 6443
      hostPort: 6443
      name: https
    - containerPort: 8080
      hostPort: 8080
      name: local
    volumeMounts:
    - mountPath: /etc/kubernetes/ssl
      name: ssl-certs-kubernetes
      readOnly: true
    - mountPath: /etc/ssl/certs
      name: ssl-certs-host
      readOnly: true
  volumes:
  - hostPath:
      path: /etc/kubernetes/ssl
    name: ssl-certs-kubernetes
  - hostPath:
      path: /usr/share/ca-certificates
    name: ssl-certs-host
`, struct {
		Image     string
		ClusterIP string
		RBAC      bool
	}{
		Image:     k.hyperkubeImage().String(),
		ClusterIP: ip.String(),
		RBAC:      !k.DisableRBAC,
	})
}

func (k *Kubelet) kubePodmasterYaml(ctx context.Context) ([]byte, error) {
	ip, err := k.ClusterIP.IP(ctx)
	if err != nil {
		return nil, err
	}
	return render(`
apiVersion: v1
kind: Pod
metadata:
  name: kube-podmaster
  namespace: kube-system
spec:
  hostNetwork: true
  containers:
  - name: controller-manager-elector
    image: {{.Image}}
    command:
    - /podmaster
    - --etcd-servers=http://{{.ClusterIP}}:2379
    - --key=controller
    - --whoami={{.ClusterIP}}
    - --source-file=/src/manifests/kube-controller-manager.yaml
    - --dest-file=/dst/manifests/kube-controller-manager.yaml
    terminationMessagePath: /dev/termination-log
    volumeMounts:
    - mountPath: /src/manifests
      name: manifest-src
      readOnly: true
    - mountPath: /dst/manifests
      name: manifest-dst
  - name: scheduler-elector
    image: {{.Image}}
    command:
    - /podmaster
    - --etcd-servers=http://{{.ClusterIP}}:2379
    - --key=scheduler
    - --whoami={{.ClusterIP}}
    - --source-file=/src/manifests/kube-scheduler.yaml
    - --dest-file=/dst/manifests/kube-scheduler.yaml
    volumeMounts:
    - mountPath: /src/manifests
      name: manifest-src
      readOnly: true
    - mountPath: /dst/manifests
      name: manifest-dst
  volumes:
  - hostPath:
      path: /srv/kubernetes/manifests
    name: manifest-src
  - hostPath:
      path: /etc/kubernetes/manifests
    name: manifest-dst
`, struct {
		Image     string
		ClusterIP string
	}{
		Image:     k.podmasterImage().String(),
		ClusterIP: ip.String(),
	})
}

func (k *Kubelet) nodeKubeconfigYaml(ctx context.Context) ([]byte, error) {
	return render(`
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    certificate-authority: /etc/kubernetes/ssl/ca.pem
    server: http://127.0.0.1:8080
users:
- name: kubelet
  user:
    client-certificate: /etc/kubernetes/ssl/node.pem
    client-key: /etc/kubernetes/ssl/node-key.pem
contexts:
- context:
    cluster: local
    user: kubelet
  name: kubelet-context
current-context: kubelet-context
`, struct{}{})
}

func (k *Kubelet) kubeProxyYaml(ctx context.Context) ([]byte, error) {
	return render(`
apiVersion: v1
kind: Pod
metadata:
  name: kube-proxy
  namespace: kube-system
spec:
  hostNetwork: true
  containers:
  - name: kube-proxy
    image: {{.Image}}
    command:
    - /hyperkube
    - proxy
    - --master=http://127.0.0.1:8080
    - --proxy-mode=iptables
    securityContext:
      privileged: true
    volumeMounts:
      - mountPath: /etc/ssl/certs
        name: ssl-certs-host
        readOnly: true
  volumes:
    - name: ssl-certs-host
      hostPath:
        path: "/usr/share/ca-certificates"
`, struct {
		Image string
	}{
		Image: k.hyperkubeImage().String(),
	})
}

func (k *Kubelet) kubeSchedulerYaml(ctx context.Context) ([]byte, error) {
	return render(`
apiVersion: v1
kind: Pod
metadata:
  name: kube-scheduler
  namespace: kube-system
spec:
  hostNetwork: true
  containers:
  - name: kube-scheduler
    image: {{.Image}}
    command:
    - /hyperkube
    - scheduler
    - --master=http://127.0.0.1:8080
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        path: /healthz
        port: 10251
      initialDelaySeconds: 15
      timeoutSeconds: 1
`, struct {
		Image string
	}{
		Image: k.hyperkubeImage().String(),
	})
}

func (k *Kubelet) kubeControllerManagerYaml(ctx context.Context) ([]byte, error) {
	return render(`
apiVersion: v1
kind: Pod
metadata:
  name: kube-controller-manager
  namespace: kube-system
spec:
  hostNetwork: true
  containers:
  - name: kube-controller-manager
    image: {{.Image}}
    command:
    - /hyperkube
    - controller-manager
    - --master=http://127.0.0.1:8080
    - --service-account-private-key-file=/etc/kubernetes/ssl/node-key.pem
    - --root-ca-file=/etc/kubernetes/ssl/ca.pem
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        path: /healthz
        port: 10252
      initialDelaySeconds: 15
      timeoutSeconds: 1
    volumeMounts:
    - mountPath: /etc/kubernetes/ssl
      name: ssl-certs-kubernetes
      readOnly: true
    - mountPath: /etc/ssl/certs
      name: ssl-certs-host
      readOnly: true
  volumes:
  - name: ssl-certs-kubernetes 
    hostPath:
      path: /etc/kubernetes/ssl
  - name: ssl-certs-host 
    hostPath:
      path: /usr/share/ca-certificates
`, struct {
		Image string
	}{
		Image: k.hyperkubeImage().String(),
	})
}

func (k *Kubelet) hyperkubeImage() docker.Image {
	return docker.Image{
		Repository: "bborbe/hyperkube",
		Tag:        k.Version,
	}
}

func (k *Kubelet) podmasterImage() docker.Image {
	return docker.Image{
		Repository: "bborbe/podmaster",
		Tag:        "1.1",
	}
}

func (k *Kubelet) pauseImage() docker.Image {
	return docker.Image{
		Repository: "bborbe/pause",
		Tag:        "3.1",
	}
}

func render(content string, data interface{}) ([]byte, error) {
	tpl, err := template.New("template").Parse(content)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	if err := tpl.Execute(b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (k *Kubelet) certDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "get homedir failed")
	}
	return fmt.Sprintf("%s/.kube/%s", usr.HomeDir, k.Context), nil
}

func (k *Kubelet) generateKeys(ctx context.Context) error {
	ip, err := k.ClusterIP.IP(ctx)
	if err != nil {
		return err
	}
	certDir, err := k.certDirectory()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return err
	}

	ca, err := triple.NewCA(fmt.Sprintf("%s-certificate-authority", k.Context))
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, caKey), cert.EncodePrivateKeyPEM(ca.Key), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, caCert), cert.EncodeCertPEM(ca.Cert), 0600); err != nil {
		return err
	}

	const name = "kubernetes"
	const namespace = "default"
	server, err := triple.NewServerKeyPair(
		ca,
		fmt.Sprintf("%s.%s.svc", name, namespace),
		name,
		namespace,
		"cluster.local",
		[]string{"10.103.0.1", ip.String()},
		[]string{},
	)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, serverKey), cert.EncodePrivateKeyPEM(server.Key), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, serverCert), cert.EncodeCertPEM(server.Cert), 0600); err != nil {
		return err
	}

	client, err := triple.NewClientKeyPair(ca, fmt.Sprintf("%s-admin", k.Context), nil)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, clientKey), cert.EncodePrivateKeyPEM(client.Key), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(certDir, clientCert), cert.EncodeCertPEM(client.Cert), 0600); err != nil {
		return err
	}

	return nil
}

func (k *Kubelet) readPem(ctx context.Context, name string) ([]byte, error) {
	certDir, err := k.certDirectory()
	if err != nil {
		return nil, err
	}
	filepath := path.Join(certDir, name)
	_, err = os.Stat(filepath)
	if os.IsNotExist(err) {
		if err := k.generateKeys(ctx); err != nil {
			return nil, err
		}
	}
	return ioutil.ReadFile(filepath)
}
