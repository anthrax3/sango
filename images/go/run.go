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

	if race, ok := in.Options["race"].(bool); ok {
		if race {
			args = append(args, "-race")
		}
	}

	return "go", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

func format(code string) string {
	v, _ := agent.System(".", code, "goimports")
	return v
}

func version() string {
	v, _ := agent.System(".", "", "go", "version")
	v = strings.Replace(v, "go version", "", -1)
	return v
}

func main() {
	agent.Run(build, run, format, version)
}
