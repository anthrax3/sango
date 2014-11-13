package main

import (
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

func build(files []string, in sango.Input, out *sango.Output) (string, []string) {
	var args []string = []string{
		"-o",
		"main",
		"-fconstant-string-class=NSConstantString",
	}

        v, _ := sango.System(".", "", "gnustep-config", "--objc-flags")
	v = strings.Replace(v, " -I/root/GNUstep/Library/Headers", "", -1)
	v = strings.Replace(v, "\n", "", -1)
	args = append(args, strings.Split(v, " ")...)

	args = append(args, files...)
	args = append(args, "-lgnustep-base")

	v, _ = sango.System(".", "", "gnustep-config", "--objc-libs")
	v = strings.Replace(v, "\n", "", -1)
	args = append(args, strings.Split(v, " ")...)

	return "gcc", args
}

func run([]string, sango.Input, *sango.Output) (string, []string) {
	return "./main", nil
}

var r = regexp.MustCompile("\\(.+\\)")

func version() string {
	_, v := sango.System(".", "", "gcc", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func main() {
	sango.Run(build, run, version)
}
