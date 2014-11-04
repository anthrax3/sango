package agent

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack"
)

type VersionHandler func() string
type CmdHandler func([]string, Input, *Output) (string, []string)

type agent struct {
	in       Input
	out      Output
	files    []string
	buildCmd CmdHandler
	runCmd   CmdHandler
}

func Run(buildCmd, runCmd CmdHandler, verCmd VersionHandler) {
	var version bool
	flag.BoolVar(&version, "v", false, "")
	flag.Parse()
	if version {
		os.Stdout.Write([]byte(strings.Trim(verCmd(), "\r\n ")))
		return
	}

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
		buildCmd: buildCmd,
		runCmd:   runCmd,
	}

	err = a.build()
	if err == nil {
		err = a.run()
	}
	if err == nil {
		a.out.Status = "Success"
	} else {
		a.out.Status = err.Error()
	}

	a.close()
}

func QuickExec(command string, args ...string) (string, string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	return string(stdout.Bytes()), string(stderr.Bytes())
}

func (a *agent) build() error {
	if a.buildCmd == nil {
		return nil
	}

	cmd, args := a.buildCmd(a.files, a.in, &a.out)
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: os.Stderr, Tag: "build-stdout"}
	msgStderr := MsgpackFilter{Writer: os.Stderr, Tag: "build-stderr"}
	err, code, signal := Exec(cmd, args, "", io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
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

func (a *agent) run() error {
	if a.runCmd == nil {
		return nil
	}

	cmd, args := a.runCmd(a.files, a.in, &a.out)
	var stdout, stderr bytes.Buffer
	msgStdout := MsgpackFilter{Writer: os.Stderr, Tag: "run-stdout"}
	msgStderr := MsgpackFilter{Writer: os.Stderr, Tag: "run-stderr"}
	start := time.Now()
	err, code, signal := Exec(cmd, args, a.in.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	a.out.RunningTime = time.Now().Sub(start).Seconds()
	a.out.RunStdout = string(stdout.Bytes())
	a.out.RunStderr = string(stderr.Bytes())
	a.out.Code = code
	a.out.Signal = signal
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
