package main

import (
	"strings"

	"github.com/h2so5/sango/src"
)

type Agent struct {
	sango.AgentBase
}

func (a Agent) BuildCommand(in sango.Input) ([]string, error) {
	var args []string = []string{
		"build",
		"-o",
		"main",
	}

	if race, ok := in.Options["race"].(bool); ok {
		if race {
			args = append(args, "-race")
		}
	}

	return append([]string{"go"}, append(args, sango.MapToFileList(in.Files)...)...), nil
}

func (a Agent) RunCommand(in sango.Input) ([]string, error) {
	return []string{"./main"}, nil
}

func (a Agent) Version() string {
	v, _ := sango.System(".", "", "go", "version")
	v = strings.Replace(v, "go version", "", -1)
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.go": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
