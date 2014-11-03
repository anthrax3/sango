package main

import "../../agent"

func build(files []string, in agent.Input, out*agent.Output) (string, []string) {
	var args []string = []string{
		"-o",
		"main",
	}
	return "g++", append(args, files...)
}

func run([]string, agent.Input, *agent.Output) (string, []string) {
	return "./main", nil
}

func main() {
	agent.Run(build, run)
}
