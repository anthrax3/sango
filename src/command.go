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
	"time"

	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"
)

type AgentCommand interface {
	Invoke() interface{}
}

type VersionHandler func() string
type CmdHandler func([]string, Input, *Output) (string, []string)
type TestHandler func() ([]string, string, string)

type agent struct {
	in       Input
	out      Output
	files    []string
	buildCmd CmdHandler
	runCmd   CmdHandler
}

type AgentOption struct {
	BuildCmd, RunCmd CmdHandler
	VerCmd           VersionHandler
	Test             TestHandler
}

func Run(opt AgentOption) {
	flag.Parse()
	subcommand := flag.Arg(0)

	switch subcommand {
	case "version":
		var img Image
		data, err := ioutil.ReadFile("config.yml")
		if err != nil {
			return
		}

		err = yaml.Unmarshal(data, &img)
		if err != nil {
			return
		}

		data, _ = ioutil.ReadFile("template.txt")
		img.Template = string(data)

		ver := strings.Trim(opt.VerCmd(), "\r\n ")
		img.Version = ver

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(img)
		os.Stdout.Close()

	case "test":
		if opt.Test != nil {
			files, stdin, stdout := opt.Test()
			a := agent{
				in:       Input{Stdin: stdin},
				files:    files,
				buildCmd: opt.BuildCmd,
				runCmd:   opt.RunCmd,
			}

			var stderr bytes.Buffer
			_, err := a.build(&stderr)
			if err != nil {
				log.Fatal(err)
			} else {
				out, err := a.run(&stderr)
				if err != nil {
					log.Fatal(err)
				}
				if out.Stdout != stdout {
					log.Fatalf("stdout should be %s; got %s", stdout, out.Stdout)
				} else {
					log.Print("TEST PASS")
				}
			}
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

		c := make(map[string]string)
		if opt.BuildCmd != nil {
			var out Output
			cmd, args := opt.BuildCmd(files, in, &out)
			c["build"] = strings.Join(append([]string{cmd}, args...), " ")
		}
		if opt.RunCmd != nil {
			var out Output
			cmd, args := opt.RunCmd(files, in, &out)
			c["run"] = strings.Join(append([]string{cmd}, args...), " ")
		}
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(c)
		os.Stdout.Close()

	case "build":
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

		a := agent{
			in:       in,
			files:    files,
			buildCmd: opt.BuildCmd,
			runCmd:   opt.RunCmd,
		}

		stage, _ := a.build(os.Stderr)
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(stage)
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

		a := agent{
			in:       in,
			files:    files,
			buildCmd: opt.BuildCmd,
			runCmd:   opt.RunCmd,
		}

		stage, _ := a.run(os.Stderr)
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(stage)
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

func (a *agent) build(msgout io.Writer) (Stage, error) {
	var s Stage
	if a.buildCmd == nil {
		return s, nil
	}

	cmd, args := a.buildCmd(a.files, a.in, &a.out)
	s.Command = strings.Join(append([]string{cmd}, args...), " ")
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: msgout, Tag: "stdout", Stage: "build"}
	msgStderr := MsgpackFilter{Writer: msgout, Tag: "stderr", Stage: "build"}
	err, code, signal := Exec(cmd, args, "", io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	s.Stdout = string(stdout.Bytes())
	s.Stderr = string(stderr.Bytes())
	s.Code = code
	s.Signal = signal
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			s.Status = "Time limit exceeded"
		} else {
			s.Status = "Failed"
			return s, errors.New(s.Status)
		}
	} else {
		s.Status = "OK"
	}

	return s, nil
}

func (a *agent) run(msgout io.Writer) (Stage, error) {
	var s Stage
	if a.runCmd == nil {
		return s, nil
	}

	cmd, args := a.runCmd(a.files, a.in, &a.out)
	s.Command = strings.Join(append([]string{cmd}, args...), " ")
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: msgout, Tag: "stdout", Stage: "run"}
	msgStderr := MsgpackFilter{Writer: msgout, Tag: "stderr", Stage: "run"}
	start := time.Now()
	err, code, signal := Exec(cmd, args, a.in.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	s.RunningTime = time.Now().Sub(start).Seconds()
	s.Stdout = string(stdout.Bytes())
	s.Stderr = string(stderr.Bytes())
	s.Code = code
	s.Signal = signal
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			s.Status = "Time limit exceeded"
		} else {
			s.Status = "Failed"
			return s, errors.New(s.Status)
		}
	} else {
		s.Status = "OK"
	}
	return s, nil
}
