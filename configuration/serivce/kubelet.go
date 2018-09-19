package service

import (
	"bytes"
	"context"
	"html/template"

	"github.com/bborbe/world/pkg/remote"

	"github.com/bborbe/world/configuration/build"
	"github.com/bborbe/world/pkg/docker"
	"github.com/bborbe/world/pkg/ssh"
	"github.com/bborbe/world/pkg/validation"
	"github.com/bborbe/world/pkg/world"
)

type Kubelet struct {
	SSH     ssh.SSH
	Version docker.Tag
}

func (k *Kubelet) Children() []world.Configuration {
	return []world.Configuration{
		&build.Hyperkube{
			Image: k.hyperkubeImage(),
		},
		&build.Podmaster{
			Image: k.podmasterImage(),
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
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/kube-apiserver.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubeApiserverYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/kube-podmaster.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.kubePodmasterYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/node-kubeconfig.yaml",
			User:    "root",
			Group:   "root",
			Perm:    0644,
			Content: remote.ContentFunc(k.nodeKubeconfigYaml),
		},
		&File{
			SSH:     k.SSH,
			Path:    "/etc/kubernetes/kube-proxy.yaml",
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
			SSH:        k.SSH,
			Name:       "kubelet",
			Memory:     2048,
			HostNet:    true,
			HostPid:    true,
			Privileged: true,
			Volumes: []string{
				"/:/rootfs:ro",
				"/sys:/sys:ro",
				"/var/log/:/var/log:rw",
				"/var/lib/docker/:/var/lib/docker:rw",
				"/var/lib/kubelet/:/var/lib/kubelet:rw,rslave",
				"/run:/run:rw",
				"/var/run:/var/run:rw",
				"/etc/kubernetes:/etc/kubernetes",
				"/srv/kubernetes:/srv/kubernetes",
			},
			Image:   k.hyperkubeImage(),
			Command: "/hyperkube",
			Args: []string{
				"kubelet",
				"--containerized",
				"--register-node=true",
				"--allow-privileged=true",
				"--pod-manifest-path=/etc/kubernetes/manifests",
				"--hostname-override=172.16.72.10",
				"--cluster-dns=10.103.0.10",
				"--cluster-domain=cluster.local",
				"--kubeconfig=/etc/kubernetes/node-kubeconfig.yaml",
				"--node-labels=etcd=true,nfsd=true,worker=true,master=true",
				"--v=0",
			},
		},
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
	)
}

func (k *Kubelet) kubeApiserverYaml() ([]byte, error) {
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
    - --etcd-servers=http://172.16.72.10:2379
    - --storage-backend=etcd3
    - --allow-privileged=true
    - --service-cluster-ip-range=10.103.0.0/16
    - --secure-port=443
    - --advertise-address=172.16.72.10
    - --admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,DefaultStorageClass,ResourceQuota
    - --tls-cert-file=/etc/kubernetes/ssl/node.pem
    - --tls-private-key-file=/etc/kubernetes/ssl/node-key.pem
    - --client-ca-file=/etc/kubernetes/ssl/ca.pem
    - --service-account-key-file=/etc/kubernetes/ssl/node-key.pem
    - --runtime-config=extensions/v1beta1/networkpolicies=true,batch/v2alpha1=true
    - --anonymous-auth=false
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        port: 8080
        path: /healthz
      initialDelaySeconds: 15
      timeoutSeconds: 15
    ports:
    - containerPort: 443
      hostPort: 443
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
		Image string
	}{
		Image: k.hyperkubeImage().String(),
	})
}

func (k *Kubelet) kubePodmasterYaml() ([]byte, error) {
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
    - --etcd-servers=http://172.16.72.10:2379
    - --key=controller
    - --whoami=172.16.72.10
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
    - --etcd-servers=http://172.16.72.10:2379
    - --key=scheduler
    - --whoami=172.16.72.10
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
		Image string
	}{
		Image: k.podmasterImage().String(),
	})
}

func (k *Kubelet) nodeKubeconfigYaml() ([]byte, error) {
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

func (k *Kubelet) kubeProxyYaml() ([]byte, error) {
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

func (k *Kubelet) kubeSchedulerYaml() ([]byte, error) {
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

func (k *Kubelet) kubeControllerManagerYaml() ([]byte, error) {
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
