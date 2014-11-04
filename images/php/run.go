package main

import (
	"regexp"
	"strings"

	"../../agent"
)

func run(files []string, in agent.Input, out *agent.Output) (string, []string) {
	return "php", files
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	v, _ := agent.QuickExec("php", "-v")
	v = strings.Split(v, "\n")[0]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	agent.Run(nil, run, version)
}
