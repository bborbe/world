# World 

My infrastructur as code

## world-validate

Validate all

```
world-validate \
-v=2 \
-logtostderr
```

## world-apply

Apply all applications

```
world-apply \
-v=2 \
-logtostderr
```

Apply all applications on cluster netcup

```
world-apply \
-v=2 \
-logtostderr \
-cluster=netcup
```

Apply application monitoring on all clusters

```
world-apply \
-v=2 \
-logtostderr \
-app=monitoring
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
