package main

import (
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

func build(files []string, in sango.Input, out *sango.Output) (string, []string) {
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

	return "gcc", append(args, files...)
}

func run([]string, sango.Input, *sango.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := sango.System(".", "", "gcc", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	sango.Run(sango.AgentOption{
		BuildCmd: build,
		RunCmd: run,
		VerCmd: version,
	})
}
