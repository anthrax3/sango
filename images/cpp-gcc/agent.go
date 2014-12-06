package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

var r = regexp.MustCompile("\\(.+\\)")

type Agent struct {
}

func (a Agent) Command(in sango.Input, n string) (string, []string, error) {
	switch n {
	case "build":
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

		return "g++", append(args, sango.MapToFileList(in.Files)...), nil

	case "run":
		if valgrind, ok := in.Options["valgrind"].(bool); ok && valgrind {
			return "valgrind", []string{"--leak-check=full", "./main"}, nil
		}
		return "./main", nil, nil
	}
	return "", nil, errors.New("unknown command")
}

func (a Agent) Version() string {
	_, v := sango.System(".", "", "g++", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.cpp": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
