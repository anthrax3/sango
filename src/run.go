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

const ProtocolVersion = 4

type AgentBase struct {
}

func (a AgentBase) BuildCommand(in Input) ([]string, error) {
	return nil, errors.New("unknown command")
}

type Agent interface {
	BuildCommand(in Input) ([]string, error)
	RunCommand(in Input) ([]string, error)
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

func Run(opt Agent) {
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

		ver := strings.Trim(opt.Version(), "\r\n ")
		img.Version = ver
		img.Protocol = ProtocolVersion
		img.Actions = []string{"run"}

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(img)
		os.Stdout.Close()

	case "test":
		files, stdin, stdout := opt.Test()
		in := Input{Stdin: stdin, Files: files}

		var stderr bytes.Buffer
		a, err := opt.BuildCommand(in)
		if err == nil {
			_, err := jtime(a, "build", in, &stderr)
			if err != nil {
				log.Fatal(err)
			}
		}
		a, err = opt.RunCommand(in)
		if err != nil {
			log.Fatal(err)
		}
		r, err := jtime(a, "run", in, &stderr)
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

		var files []string
		for k, v := range in.Files {
			ioutil.WriteFile(k, []byte(v), 0644)
			files = append(files, k)
		}

		var command = map[string]string{}

		a, err := opt.BuildCommand(in)
		if err == nil {
			command["build"] = strings.Join(a, " ")
		}
		a, err = opt.RunCommand(in)
		if err == nil {
			command["run"] = strings.Join(a, " ")
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

		var files []string
		for k, v := range in.Files {
			ioutil.WriteFile(k, []byte(v), 0644)
			files = append(files, k)
		}

		out := Output{Results: make(map[string]ExecResult)}
		out.Status = "Success"

		var builderr bool
		a, err := opt.BuildCommand(in)
		if err == nil {
			r, err := jtime(a, "build", in, os.Stderr)
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
			a, err = opt.RunCommand(in)
			if err != nil {
				log.Fatal(err)
			}
			r, err := jtime(a, "run", in, os.Stderr)
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

func jtime(a []string, p string, in Input, msgout io.Writer) (ExecResult, error) {
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
