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
