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

		return "clang++", append(args, sango.MapToFileList(in.Files)...), nil

	case "run":
		return "./main", nil, nil
	}
	return "", nil, errors.New("unknown command")
}

func (a Agent) Version() string {
	_, v := sango.System(".", "", "clang++", "-v")
	l := strings.Split(v, "\n")
	v = l[0]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	v = strings.Replace(v, "Ubuntu", "", -1)
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.cpp": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
