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
	}

	if optim, ok := in.Options["optim"].(string); ok {
		switch optim {
			case "-O0":
				args = append(args, optim)
			case "-O1":
				args = append(args, optim)
			case "-O2":
				args = append(args, optim)
			case "-O3":
				args = append(args, optim)
			case "-Os":
				args = append(args, optim)
		}
	}

	return "g++", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := agent.System(".", "g++", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	agent.Run(build, run, version)
}
