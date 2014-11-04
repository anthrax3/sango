package main

import (
	"strings"

	"../../agent"
)

func build(files []string, in agent.Input, out *agent.Output) (string, []string) {
	var args []string = []string{
		"build",
		"-o",
		"main",
	}
	return "go", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

func version() string {
	v, _ := agent.System(".", "go", "version")
	v = strings.Replace(v, "go version", "", -1)
	return v
}

func main() {
	agent.Run(build, run, version)
}
