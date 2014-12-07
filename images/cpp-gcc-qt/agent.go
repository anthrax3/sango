package main

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/h2so5/sango/src"
)

var r = regexp.MustCompile("\\(.+\\)")

type Agent struct {
	sango.AgentBase
}

func (a Agent) BuildCommand(in sango.Input) ([]string, error) {
	return []string{
		"sh",
		"-c",
		"qmake -project && qmake && make",
	}, nil
}

func (a Agent) RunCommand(in sango.Input) ([]string, error) {
	return []string{"./sango"}, nil
}

func (a Agent) Version() string {
	_, v := sango.System(".", "", "g++", "-v")
	l := strings.Split(v, "\n")
	v = l[len(l)-2]
	v = string(r.ReplaceAll([]byte(v), []byte("")))

	q, _ := sango.System(".", "", "qmake", "-query", "QT_VERSION")
	v += " + Qt" + q
	return v
}

func (a Agent) Test() (map[string]string, string, string) {
	return map[string]string{"test/hello.cpp": ""}, "", "Hello World"
}

func (a Agent) ActionCommands(in sango.Input) (map[string][]string, error) {
	return map[string][]string{
		"fmt": append([]string{"goimports", "-w"}, sango.MapToFileList(in.Files)...),
	}, nil
}

func (a Agent) Action(c string, in sango.Input) (sango.ExecResult, error) {
	if c == "fmt" {
		a, err := a.ActionCommands(in)
		if err != nil {
			return sango.ExecResult{}, err
		}
		r, err := sango.Jtime(a["fmt"], "fmt", in, nil)
		files := map[string]string{}
		for k := range in.Files {
			data, err := ioutil.ReadFile(k)
			if err == nil {
				files[k] = string(data)
			}
		}
		r.Data = files
		return r, err
	}
	return sango.ExecResult{}, errors.New("unknown command")
}

func main() {
	sango.Run(Agent{})
}
