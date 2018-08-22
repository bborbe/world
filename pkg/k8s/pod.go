package k8s

type PodSpec struct {
	Containers  []PodContainer `yaml:"containers"`
	Volumes     []PodVolume    `yaml:"volumes,omitempty"`
	HostNetwork PodHostNetwork `yaml:"hostNetwork,omitempty"`
	DnsPolicy   PodDnsPolicy   `yaml:"dnsPolicy,omitempty"`
}

type PodVolumeName string

type PodNfsPath string

type PodNfsServer string

type PodVolumeNfs struct {
	Path   PodNfsPath   `yaml:"path"`
	Server PodNfsServer `yaml:"server"`
}

type PodVolumeSecret struct {
	Name  PodSecretName   `yaml:"secretName"`
	Items []PodSecretItem `yaml:"items"`
}

type PodSecretName string

type PodSecretItem struct {
	Key  PodSecretItemKey  `yaml:"key"`
	Path PodSecretItemPath `yaml:"path"`
}

type PodSecretItemKey string

type PodSecretItemPath string

type PodVolumeConfigMap struct {
	Name  PodConfigMapName   `yaml:"name"`
	Items []PodConfigMapItem `yaml:"items"`
}

type PodConfigMapName string

type PodConfigMapItem struct {
	Key  PodConfigMapItemKey  `yaml:"key"`
	Path PodConfigMapItemPath `yaml:"path"`
}

type PodConfigMapItemKey string

type PodConfigMapItemPath string

type PodVolumeEmptyDir struct{}

type PodVolume struct {
	Name      PodVolumeName      `yaml:"name"`
	Nfs       PodVolumeNfs       `yaml:"nfs,omitempty"`
	ConfigMap PodVolumeConfigMap `yaml:"configMap,omitempty"`
	Secret    PodVolumeSecret    `yaml:"secret,omitempty"`
	EmptyDir  *PodVolumeEmptyDir `yaml:"emptyDir,omitempty"`
}

type PodDnsPolicy string

type PodHostNetwork bool

type Arg string

type Env struct {
	Name      string    `yaml:"name"`
	Value     string    `yaml:"value,omitempty"`
	ValueFrom ValueFrom `yaml:"valueFrom,omitempty"`
}

type ValueFrom struct {
	SecretKeyRef SecretKeyRef `yaml:"secretKeyRef"`
}

type SecretKeyRef struct {
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}

type PodImage string

type PodName string

type PodPortContainerPort int

type PodPortHostPort int

type PodPortName string

type PodPortProtocol string

type PodPort struct {
	ContainerPort PodPortContainerPort `yaml:"containerPort,omitempty"`
	HostPort      PodPortHostPort      `yaml:"hostPort,omitempty"`
	Name          PodPortName          `yaml:"name,omitempty"`
	Protocol      PodPortProtocol      `yaml:"protocol,omitempty"`
}

type CpuLimit string
type MemoryLimit string

type Resources struct {
	Cpu    string      `yaml:"cpu"`
	Memory MemoryLimit `yaml:"memory"`
}

type PodResources struct {
	Limits   Resources `yaml:"limits"`
	Requests Resources `yaml:"requests"`
}

type VolumeMountPath string

type VolumeName string

type VolumeReadOnly bool

type VolumeMount struct {
	Path     VolumeMountPath `yaml:"mountPath"`
	Name     VolumeName      `yaml:"name"`
	ReadOnly VolumeReadOnly  `yaml:"readOnly"`
}

type PodContainer struct {
	Args         []Arg         `yaml:"args,omitempty"`
	Env          []Env         `yaml:"env,omitempty"`
	Image        PodImage      `yaml:"image"`
	Name         PodName       `yaml:"name"`
	Ports        []PodPort     `yaml:"ports,omitempty"`
	Resources    PodResources  `yaml:"resources"`
	VolumeMounts []VolumeMount `yaml:"volumeMounts,omitempty"`
}
