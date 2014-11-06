package main

import (
	"regexp"
	"strings"

	"../../src"
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

	return "clang++", append(args, files...)
}

func run([]string, sango.Input, *sango.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := sango.System(".", "", "clang++", "-v")
	l := strings.Split(v, "\n")
	v = l[0]
        v = string(r.ReplaceAll([]byte(v), []byte("")))
        v = strings.Replace(v, "Ubuntu", "", -1)
	return v
}

func main() {
	sango.Run(build, run, version)
}
