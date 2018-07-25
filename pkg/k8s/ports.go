package k8s

type PortName string
type PortNumber int
type PortProtocol string
type PortTarget string

type Port struct {
	Name       PortName     `yaml:"name"`
	Port       PortNumber   `yaml:"port"`
	Protocol   PortProtocol `yaml:"protocol"`
	TargetPort PortTarget   `yaml:"targetPort"`
}
