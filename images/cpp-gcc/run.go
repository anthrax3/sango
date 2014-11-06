package main

import (
	"regexp"
	"strings"

	"../../agent"
)

func build(files []string, in agent.Input, out *agent.Output) (string, []string) {
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

	return "g++", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := agent.System(".", "", "g++", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	agent.Run(build, run, version)
}
