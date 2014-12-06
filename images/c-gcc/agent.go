package main

import (
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

var r = regexp.MustCompile("\\(.+\\)")

type Agent struct {
	sango.AgentBase
}

func (a Agent) BuildCommand(in sango.Input) ([]string, error) {

	var args []string = []string{
		"-o",
		"main",
		"-pthread",
	}

	if optim, ok := in.Options["optim"].(string); ok {
		args = append(args, optim)
	}

	if optim, ok := in.Options["std"].(string); ok {
		args = append(args, optim)
	}

	return append([]string{"gcc"}, append(args, sango.MapToFileList(in.Files)...)...), nil

}

func (a Agent) RunCommand(in sango.Input) ([]string, error) {
	if valgrind, ok := in.Options["valgrind"].(bool); ok && valgrind {
		return []string{"valgrind", "--leak-check=full", "./main"}, nil
	}
	return []string{"./main"}, nil
}

func (a Agent) Version() string {
	_, v := sango.System(".", "", "gcc", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.c": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
