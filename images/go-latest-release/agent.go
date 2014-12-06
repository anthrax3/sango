package main

import (
	"errors"
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

func (a Agent) ActionCommands(in sango.Input) (map[string][]string, error) {
	return map[string][]string{
		"fmt": append([]string{"gofmt"}, sango.MapToFileList(in.Files)...),
	}, nil
}

func (a Agent) Action(c string, in sango.Input) (sango.ExecResult, error) {
	if c == "fmt" {
		a, err := a.ActionCommands(in)
		if err != nil {
			return sango.ExecResult{}, err
		}
		return sango.Jtime(a["fmt"], "fmt", in, nil)
	}
	return sango.ExecResult{}, errors.New("unknown command")
}

func main() {
	sango.Run(Agent{})
}
