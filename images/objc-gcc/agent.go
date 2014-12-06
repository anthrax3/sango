package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

var r = regexp.MustCompile("\\(.+\\)")

type Agent struct {
}

func (a Agent) Command(in sango.Input, n string) (string, []string, error) {
	switch n {
	case "build":
		var args []string = []string{
			"-o",
			"main",
			"-fconstant-string-class=NSConstantString",
		}

		v, _ := sango.System(".", "", "gnustep-config", "--objc-flags")
		v = strings.Replace(v, " -I/root/GNUstep/Library/Headers", "", -1)
		v = strings.Replace(v, "\n", "", -1)
		args = append(args, strings.Split(v, " ")...)

		args = append(args, sango.MapToFileList(in.Files)...)
		args = append(args, "-lgnustep-base")

		v, _ = sango.System(".", "", "gnustep-config", "--objc-libs")
		v = strings.Replace(v, "\n", "", -1)
		args = append(args, strings.Split(v, " ")...)

		return "gcc", args, nil

	case "run":
		return "./main", nil, nil
	}
	return "", nil, errors.New("unknown command")
}

func (a Agent) Version() string {
	_, v := sango.System(".", "", "gcc", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.m": ""}, "", "Hello World"
}

func main() {
	sango.Run(Agent{})
}
