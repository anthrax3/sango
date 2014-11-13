package main

import (
	"strings"

	"github.com/h2so5/sango/src"
)

func build(files []string, in sango.Input, out *sango.Output) (string, []string) {
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

func run([]string, sango.Input, *sango.Output) (string, []string) {
	return "./main", nil
}

func version() string {
	v, _ := sango.System(".", "", "go", "version")
	v = strings.Replace(v, "go version", "", -1)
	return v
}

func main() {
	sango.Run(sango.AgentOption{
		BuildCmd: build,
		RunCmd: run,
		VerCmd: version,
	})
}
