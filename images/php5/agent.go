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
	case "run":
		return "php", sango.MapToFileList(in.Files), nil
	}
	return "", nil, errors.New("unknown command")
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
