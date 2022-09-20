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

Apply all applications on cluster fire

```
world apply \
-v=2 \
-cluster=fire
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
