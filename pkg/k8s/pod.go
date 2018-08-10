package k8s

type PodSpec struct {
	Containers  []PodContainer `yaml:"containers"`
	Volumes     []PodVolume    `yaml:"volumes"`
	HostNetwork PodHostNetwork `yaml:"hostNetwork"`
}

type PodVolumeName string

type PodNfsPath string

type PodNfsServer string

type PodNfs struct {
	Path   PodNfsPath   `yaml:"path"`
	Server PodNfsServer `yaml:"server"`
}

type EmptyDir map[string]string

type PodVolume struct {
	Name     PodVolumeName `yaml:"name"`
	Nfs      PodNfs        `yaml:"nfs,omitempty"`
	EmptyDir EmptyDir      `yaml:"emptyDir,omitempty"`
}

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

type ResourceList map[string]string

type PodResources struct {
	Limits   ResourceList `yaml:"limits"`
	Requests ResourceList `yaml:"requests"`
}

type VolumeMountPath string

type VolumeName string

type VolumeReadOnly bool

type VolumeMount struct {
	MountPath VolumeMountPath `yaml:"mountPath"`
	Name      VolumeName      `yaml:"name"`
	ReadOnly  VolumeReadOnly  `yaml:"readOnly"`
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
