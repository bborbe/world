package cluster

import "github.com/bborbe/world/pkg/k8s"

type Cluster struct {
	Context   k8s.Context
	NfsServer k8s.PodNfsServer
}
