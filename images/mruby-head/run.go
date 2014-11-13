package main

import"github.com/h2so5/sango/src"

func run(files []string, in sango.Input, out *sango.Output) (string, []string) {
	return "mruby", files
}

func version() string {
	v, _ := sango.System(".", "", "mruby", "--version")
	g, _ := sango.System("/mruby", "", "git", "rev-parse", "HEAD")
	return v[:len(v)-1] + g[:8]
}

func main() {
	sango.Run(sango.AgentOption{
		RunCmd: run,
		VerCmd: version,
	})
}
