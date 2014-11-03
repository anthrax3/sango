package main

import "../../agent"

func run(files []string, in agent.Input, out *agent.Output) (string, []string) {
	return "php", files
}

func main() {
	agent.Run(nil, run)
}
