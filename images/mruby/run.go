package main

import "../../agent"

func run(files []string, in agent.Input, out *agent.Output) (string, []string) {
	return "mruby", files
}

func version() string {
	v, _ := agent.System(".", "", "mruby", "--version")
	g, _ := agent.System("/mruby", "", "git", "rev-parse", "HEAD")
	return v[:len(v)-1] + g[:8]
}

func main() {
	agent.Run(nil, run, version)
}
