package sango

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"
)

const ProtocolVersion = 5

type AgentBase struct {
}

func (a AgentBase) BuildCommand(in Input) ([]string, error) {
	return nil, errors.New("unknown command")
}

func (a AgentBase) ActionCommands(in Input) (map[string][]string, error) {
	return nil, nil
}

func (a AgentBase) Action(c string, in Input) (ExecResult, error) {
	return ExecResult{}, errors.New("unknown command")
}

type Agent interface {
	BuildCommand(in Input) ([]string, error)
	RunCommand(in Input) ([]string, error)
	ActionCommands(in Input) (map[string][]string, error)
	Action(c string, in Input) (ExecResult, error)
	Version() string
	Test() (map[string]string, string, string)
}

func MapToFileList(files map[string]string) []string {
	l := make([]string, 0, len(files))
	for k := range files {
		l = append(l, k)
	}
	return l
}

func Run(act Agent) {
	flag.Parse()
	subcommand := flag.Arg(0)

	switch subcommand {
	case "version":
		var img Image
		data, err := ioutil.ReadFile("/tmp/sango/config.yml")
		if err != nil {
			return
		}

		err = yaml.Unmarshal(data, &img)
		if err != nil {
			return
		}

		data, _ = ioutil.ReadFile("/tmp/sango/template.txt")
		img.Template = string(data)

		ver := strings.Trim(act.Version(), "\r\n ")
		img.Version = ver
		img.Protocol = ProtocolVersion
		img.Actions = []string{"run"}

		c, err := act.ActionCommands(Input{})
		if err == nil {
			for k := range c {
				img.Actions = append(img.Actions, k)
			}
		}

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(img)
		os.Stdout.Close()

	case "test":
		files, stdin, stdout := act.Test()
		in := Input{Stdin: stdin, Files: files}

		a, err := act.BuildCommand(in)
		if err == nil {
			_, err := Jtime(a, "build", in, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
		a, err = act.RunCommand(in)
		if err != nil {
			log.Fatal(err)
		}
		r, err := Jtime(a, "run", in, nil)
		if err != nil {
			log.Fatal(err)
		}
		if r.Stdout != stdout {
			log.Fatalf("stdout should be %s; got %s", stdout, r.Stdout)
		}

	case "cmd":
		var in Input
		d := msgpack.NewDecoder(os.Stdin)
		err := d.Decode(&in)
		if err != nil {
			return
		}

		for k, v := range in.Files {
			ioutil.WriteFile(k, []byte(v), 0644)
		}

		var command = map[string]string{}

		a, err := act.BuildCommand(in)
		if err == nil {
			command["build"] = strings.Join(a, " ")
		}
		a, err = act.RunCommand(in)
		if err == nil {
			command["run"] = strings.Join(a, " ")
		}

		c, err := act.ActionCommands(in)
		if err == nil {
			for k, a := range c {
				command[k] = strings.Join(a, " ")
			}
		}

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(command)
		os.Stdout.Close()

	case "run":
		var in Input
		d := msgpack.NewDecoder(os.Stdin)
		err := d.Decode(&in)
		if err != nil {
			return
		}

		for k, v := range in.Files {
			ioutil.WriteFile(k, []byte(v), 0644)
		}

		out := Output{Results: make(map[string]ExecResult)}
		out.Status = "Success"

		var builderr bool
		a, err := act.BuildCommand(in)
		if err == nil {
			r, err := Jtime(a, "build", in, os.Stderr)
			if err != nil {
				builderr = true
				if _, ok := err.(TimeoutError); ok {
					out.Status = "Time limit exceeded"
				} else {
					out.Status = "Build error"
				}
			}
			out.Results["build"] = r
		}

		if !builderr {
			a, err = act.RunCommand(in)
			if err != nil {
				log.Fatal(err)
			}
			r, err := Jtime(a, "run", in, os.Stderr)
			if err != nil {
				if _, ok := err.(TimeoutError); ok {
					out.Status = "Time limit exceeded"
				} else {
					out.Status = "Runtime error"
				}
			}
			out.Results["run"] = r
		}

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(out)
		os.Stdout.Close()

	default:
		var in Input
		d := msgpack.NewDecoder(os.Stdin)
		err := d.Decode(&in)
		if err != nil {
			return
		}

		for k, v := range in.Files {
			ioutil.WriteFile(k, []byte(v), 0644)
		}

		out := Output{Results: make(map[string]ExecResult)}
		out.Status = "Success"

		r, err := act.Action(subcommand, in)
		if err != nil {
			out.Status = "Runtime error"
		}
		out.Results[subcommand] = r
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(out)
		os.Stdout.Close()
	}
}

func System(wdir, stdin, command string, args ...string) (string, string) {
	path, _ := os.Getwd()
	os.Chdir(wdir)
	defer os.Chdir(path)
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	return string(stdout.Bytes()), string(stderr.Bytes())
}

func Jtime(a []string, p string, in Input, msgout io.Writer) (ExecResult, error) {
	var stdout bytes.Buffer
	var result ExecResult
	cmd := exec.Command("jtime", append([]string{"-p=" + p + "-", "--"}, a...)...)
	cmd.Stdin = strings.NewReader(in.Stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = msgout
	cmd.Run()

	err := msgpack.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return ExecResult{}, err
	}

	if result.Timeout {
		return result, TimeoutError{}
	} else if result.Code != 0 {
		return result, errors.New("Runtime error")
	}
	return result, nil
}
