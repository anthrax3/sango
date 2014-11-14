package main

import (
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

func build(files []string, in sango.Input, out *sango.Output) (string, []string) {
	var args []string = []string{
		"-c",
		"qmake -project && qmake && make",
	}

	return "sh", args
}

func run([]string, sango.Input, *sango.Output) (string, []string) {
	return "./sango", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := sango.System(".", "", "g++", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))

	q, _ := sango.System(".", "", "qmake", "-query", "QT_VERSION")
	v += " + Qt" + q
	return v
}

func test() ([]string, string, string) {
	return []string{"test/hello.cpp"}, "", "Hello World"
}

func main() {
	sango.Run(sango.AgentOption{
		BuildCmd: build,
		RunCmd:   run,
		VerCmd:   version,
		Test:     test,
	})
}
