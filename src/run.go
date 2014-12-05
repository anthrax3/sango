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
	"syscall"
	"time"

	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"
)

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
			err := a.build(&stderr)
			if err == nil {
				err = a.run(&stderr)
			}
			if err != nil {
				log.Fatal(err)
			}

			if a.out.RunStdout != stdout {
				log.Fatalf("stdout should be %s; got %s", stdout, a.out.RunStdout)
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

		var c CommandLine
		if opt.BuildCmd != nil {
			var out Output
			cmd, args := opt.BuildCmd(files, in, &out)
			c.Build = strings.Join(append([]string{cmd}, args...), " ")
		}
		if opt.RunCmd != nil {
			var out Output
			cmd, args := opt.RunCmd(files, in, &out)
			c.Run = strings.Join(append([]string{cmd}, args...), " ")
		}
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(c)
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

		err = a.build(os.Stderr)
		if err == nil {
			err = a.run(os.Stderr)
		}
		if err == nil {
			a.out.Status = "Success"
		} else {
			a.out.Status = err.Error()
		}
		a.close()
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

func (a *agent) build(msgout io.Writer) error {
	if a.buildCmd == nil {
		return nil
	}

	cmd, args := a.buildCmd(a.files, a.in, &a.out)
	a.out.Command.Build = strings.Join(append([]string{cmd}, args...), " ")
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: msgout, Tag: "build-stdout"}
	msgStderr := MsgpackFilter{Writer: msgout, Tag: "build-stderr"}
	err, code, signal := Exec(cmd, args, strings.NewReader(""), io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	a.out.BuildStdout = string(stdout.Bytes())
	a.out.BuildStderr = string(stderr.Bytes())
	a.out.Code = code
	a.out.Signal = signal
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			return errors.New("Time limit exceeded")
		} else {
			return errors.New("Build error")
		}
	}

	return nil
}

func (a *agent) run(msgout io.Writer) error {
	if a.runCmd == nil {
		return nil
	}

	cmd, args := a.runCmd(a.files, a.in, &a.out)
	a.out.Command.Run = strings.Join(append([]string{cmd}, args...), " ")
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: msgout, Tag: "run-stdout"}
	msgStderr := MsgpackFilter{Writer: msgout, Tag: "run-stderr"}
	start := time.Now()
	err, code, signal := Exec(cmd, args, strings.NewReader(a.in.Stdin), io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	a.out.RunningTime = time.Now().Sub(start).Seconds()
	a.out.RunStdout = string(stdout.Bytes())
	a.out.RunStderr = string(stderr.Bytes())
	a.out.Code = code
	a.out.Signal = signal

	var usage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, &usage)
	u := Rusage{
		Utime:   float64(usage.Utime.Nano()) / 1000000000.0,
		Stime:   float64(usage.Stime.Nano()) / 1000000000.0,
		Maxrss:  usage.Maxrss,
		Minflt:  usage.Minflt,
		Majflt:  usage.Majflt,
		Inblock: usage.Inblock,
		Oublock: usage.Oublock,
		Nvcsw:   usage.Nvcsw,
		Nivcsw:  usage.Nivcsw,
	}
	a.out.Rusage = u

	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			return errors.New("Time limit exceeded")
		} else {
			return errors.New("Runtime error")
		}
	}
	return nil
}

func (a agent) close() {
	e := msgpack.NewEncoder(os.Stdout)
	e.Encode(a.out)
	os.Stdout.Close()
}
