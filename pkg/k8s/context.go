package k8s

type Context string

func (c Context) String() string {
	return string(c)
}
