# World 

My infrastructur as code

## validate

Validate all

```
world validate \
-v=2 
```

## apply

Apply all applications

```
world apply \
-v=2 
```

Apply all applications on cluster netcup

```
world apply \
-v=2 \
-cluster=netcup
```

Apply application monitoring on all clusters

```
world apply \
-v=2 \
-app=monitoring
```

## yaml-to-struct

```
world yaml-to-struct \
-v=2 \
-file=my.yaml
```

## Known Bugs

### SSL cert expired 

`Unable to connect to the server: x509: certificate has expired or is not yet valid`

```
rm -rf ~/.kube/fire
ssh fire.hm.benjamin-borbe.de
systemctl stop docker
rm -rf /srv/kubernetes /etc/kubernetes /var/lib/etcd /var/lib/kubelet /var/lib/docker
exit
world apply -c fire -a cluster -v=1
world apply -c fire -a cluster-admin -v=1
world apply -c fire -a calico -v=1
world apply -c fire -v=1

// maybe
DisableCNI:  true,
DisableRBAC: true,
systemctl restart kubelet
```

## Cert Manager 

https://github.com/jetstack/cert-manager/releases/download/v0.13.1/cert-manager.yaml
```
kubectl apply --validate=false -f cert-manger
```
