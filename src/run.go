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

const ProtocolVersion = 3

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

		img.Protocol = ProtocolVersion

		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(img)
		os.Stdout.Close()

	case "test":
		if opt.Test != nil {
			files, stdin, stdout := opt.Test()
			a := agent{
				in:       Input{Stdin: stdin},
				out:      Output{Results: make(map[string]ExecResult)},
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

			if a.out.Results["run"].Stdout != stdout {
				log.Fatalf("stdout should be %s; got %s", stdout, a.out.Results["run"].Stdout)
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

		var c = map[string]string{}
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
			out:      Output{Results: make(map[string]ExecResult)},
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

	var stdout bytes.Buffer
	var result ExecResult
	c, args := a.buildCmd(a.files, a.in, &a.out)
	cmd := exec.Command("jtime", append([]string{"-p=build-", "--", c}, args...)...)
	cmd.Stdin = strings.NewReader(a.in.Stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = msgout
	cmd.Run()

	err := msgpack.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return err
	}
	a.out.Results["build"] = result

	if result.Timeout {
		return errors.New("Time limit exceeded")
	} else if result.Code != 0 {
		return errors.New("Build error")
	}
	return nil
}

func (a *agent) run(msgout io.Writer) error {
	if a.runCmd == nil {
		return nil
	}

	var stdout bytes.Buffer
	var result ExecResult
	c, args := a.runCmd(a.files, a.in, &a.out)
	cmd := exec.Command("jtime", append([]string{"-p=run-", "--", c}, args...)...)
	cmd.Stdin = strings.NewReader(a.in.Stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = msgout
	cmd.Run()

	err := msgpack.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return err
	}
	a.out.Results["run"] = result

	if result.Timeout {
		return errors.New("Time limit exceeded")
	} else if result.Code != 0 {
		return errors.New("Runtime error")
	}
	return nil
}

func (a agent) close() {
	e := msgpack.NewEncoder(os.Stdout)
	e.Encode(a.out)
	os.Stdout.Close()
}
