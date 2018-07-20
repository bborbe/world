package world

import (
	"fmt"
)

var BuildInstructions = []BuildInstruction{
	{
		Name:            "ip",
		GitRepo:         "https://github.com/bborbe/ip.git",
		SourceDirectory: "github.com/bborbe/ip",
		Package:         "github.com/bborbe/ip/cmd/ip-server",
		Registry:        "docker.io",
		Image:           "bborbe/ip",
	},
}

type BuildInstruction struct {
	Registry        Registry
	Image           Image
	Version         Version
	Name            Name
	SourceDirectory SourceDirectory
	GitRepo         GitRepo
	Package         Package
}

func BuildInstructionForName(name Name) (*BuildInstruction, error) {
	for _, build := range BuildInstructions {
		if name == build.Name {
			return &build, nil
		}
	}
	return nil, fmt.Errorf("no build with name %s found", name)
}
