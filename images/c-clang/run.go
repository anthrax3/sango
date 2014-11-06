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

	return "clang", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	v, _ := agent.System(".", "", "clang", "--version")
	l := strings.Split(v, "\n")
	v = l[0]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	v = strings.Replace(v, "Ubuntu", "", -1)
	return v
}

func main() {
	agent.Run(build, run, version)
}
