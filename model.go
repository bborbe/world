package world

type Registry string

func (r Registry) String() string {
	return string(r)
}

type Image string

func (i Image) String() string {
	return string(i)
}

type Version string

func (v Version) String() string {
	return string(v)
}

type SourceDirectory string

func (s SourceDirectory) String() string {
	return string(s)
}

type GitRepo string

func (g GitRepo) String() string {
	return string(g)
}

type Package string

func (p Package) String() string {
	return string(p)
}

type Context string

func (c Context) String() string {
	return string(c)
}

type Name string

func (n Name) String() string {
	return string(n)
}

type Domain string

func (d Domain) String() string {
	return string(d)
}
