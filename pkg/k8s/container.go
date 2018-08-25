package k8s

type Image string

type ContainerName string

type ContainerPortNumer int

type ContainerPortHostPort int

type ContainerPortName string

type ContainerPortProtocol string

type CpuLimit string

type MemoryLimit string

type ContainerResource struct {
	Cpu    string      `yaml:"cpu"`
	Memory MemoryLimit `yaml:"memory"`
}

type Resources struct {
	Limits   ContainerResource `yaml:"limits"`
	Requests ContainerResource `yaml:"requests"`
}

type ContainerPort struct {
	ContainerPort ContainerPortNumer    `yaml:"containerPort,omitempty"`
	HostPort      ContainerPortHostPort `yaml:"hostPort,omitempty"`
	Name          ContainerPortName     `yaml:"name,omitempty"`
	Protocol      ContainerPortProtocol `yaml:"protocol,omitempty"`
}

type ContainerMountPath string

type ContainerMountName string

type ContainerMountReadOnly bool

type ContainerMount struct {
	Path     ContainerMountPath     `yaml:"mountPath"`
	Name     ContainerMountName     `yaml:"name"`
	ReadOnly ContainerMountReadOnly `yaml:"readOnly"`
}

type Arg string

type Env struct {
	Name      string    `yaml:"name"`
	Value     string    `yaml:"value,omitempty"`
	ValueFrom ValueFrom `yaml:"valueFrom,omitempty"`
}

type Container struct {
	Name           ContainerName    `yaml:"name"`
	Image          Image            `yaml:"image"`
	Args           []Arg            `yaml:"args,omitempty"`
	Env            []Env            `yaml:"env,omitempty"`
	Ports          []ContainerPort  `yaml:"ports,omitempty"`
	Resources      Resources        `yaml:"resources,omitempty"`
	VolumeMounts   []ContainerMount `yaml:"volumeMounts,omitempty"`
	ReadinessProbe Probe            `yaml:"readinessProbe,omitempty"`
	LivenessProbe  Probe            `yaml:"livenessProbe,omitempty"`
}

type Probe struct {
	HttpGet             HttpGet `yaml:"httpGet,omitempty"`
	InitialDelaySeconds int     `yaml:"initialDelaySeconds,omitempty"`
	SuccessThreshold    int     `yaml:"successThreshold,omitempty"`
	FailureThreshold    int     `yaml:"failureThreshold,omitempty"`
	TimeoutSeconds      int     `yaml:"timeoutSeconds,omitempty"`
}

type HttpGet struct {
	Path   string `yaml:"path,omitempty"`
	Port   int    `yaml:"port,omitempty"`
	Scheme string `yaml:"scheme,omitempty"`
}
