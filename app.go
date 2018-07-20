package world

import "fmt"

var Apps = []App{
	{
		Name:    "ip",
		Version: "1.1.0",
		Context: "netcup",
		Domains: []Domain{
			"ip.benjamin-borbe.de",
		},
	},
}

type App struct {
	Context Context
	Name    Name
	Version Version
	Domains []Domain
}


func GetApp(name Name) (*App, error) {
	for _, app := range Apps {
		if app.Name == name {
			return &app, nil
		}
	}
	return nil, fmt.Errorf("no app with name %s found", name.String())
}
