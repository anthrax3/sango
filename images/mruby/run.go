package main

import "../../agent"

func run(files []string, in agent.Input, out *agent.Output) (string, []string) {
	return "/mruby/build/host/bin/mruby", files
}

func version() string {
	v, _ := agent.QuickExec("/mruby/build/host/bin/mruby", "--version")
	return v
}

func main() {
	agent.Run(nil, run, version)
}
