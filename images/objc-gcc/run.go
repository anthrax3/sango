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
		"-fconstant-string-class=NSConstantString",
	}

        v, _ := agent.System(".", "", "gnustep-config", "--objc-flags")
	v = strings.Replace(v, " -I/root/GNUstep/Library/Headers", "", -1)
	v = strings.Replace(v, "\n", "", -1)
	args = append(args, strings.Split(v, " ")...)

	args = append(args, files...)
	args = append(args, "-lgnustep-base")

	v, _ = agent.System(".", "", "gnustep-config", "--objc-libs")
	v = strings.Replace(v, "\n", "", -1)
	args = append(args, strings.Split(v, " ")...)

	return "gcc", args
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := agent.System(".", "", "gcc", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	agent.Run(build, run, version)
}
