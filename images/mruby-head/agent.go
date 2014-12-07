package main

import "github.com/h2so5/sango/src"

type Agent struct {
	sango.AgentBase
}

func (a Agent) RunCommand(in sango.Input) ([]string, error) {
	return append([]string{"mruby"}, sango.MapToFileList(in.Files)...), nil
}

func (a Agent) Version() string {
	v, _ := sango.System(".", "", "mruby", "--version")
	g, _ := sango.System("/mruby", "", "git", "rev-parse", "HEAD")
	return v[:len(v)-1] + g[:8]
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.rb": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
