package main

import (
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

var r = regexp.MustCompile("\\(.+\\)")

type Agent struct {
	sango.AgentBase
}

func (a Agent) RunCommand(in sango.Input) (string, []string, error) {
	return "php", sango.MapToFileList(in.Files), nil
}

func (a Agent) Version() string {
	v, _ := sango.System(".", "", "php", "-v")
	v = strings.Split(v, "\n")[0]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.php": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
