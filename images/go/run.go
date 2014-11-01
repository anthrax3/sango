package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"time"
	"flag"

	sango "../../src"

	"github.com/vmihailenco/msgpack"
	)

	var version *bool = flag.Bool("v", false, "")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if (*version) {
		v, err := ioutil.ReadFile("/sango/version")
		if err == nil {
			os.Stdout.Write(v)
		}
		return
	}

	var out sango.Output
	defer func() {
		e := msgpack.NewEncoder(os.Stdout)
		e.Encode(out)
	}()

	var in sango.Input
	d := msgpack.NewDecoder(os.Stdin)
	err := d.Decode(&in)
	if err != nil {
		return
	}

	if len(in.Files) == 0 {
		return
	}

	var args []string = []string{
		"build",
		"-o",
		"main",
	}
	for k, v := range in.Files {
		err := ioutil.WriteFile(k, []byte(v), 0644)
		if err != nil {
			return
		}
		args = append(args, k)
	}

	stdout, stderr, err, code, signal := sango.Exec("go", args, "", 5*time.Second)
	out.BuildStdout = stdout
	out.BuildStderr = stderr
	out.Code = code
	out.Signal = signal
	if err != nil {
		if _, ok := err.(sango.TimeoutError); ok {
			out.Status = "Time limit exceeded"
		} else {
			out.Status = "Build error"
		}
		return
	}

	start := time.Now()
	stdout, stderr, err, code, signal = sango.Exec("./main", nil, in.Stdin, 5*time.Second)
	out.RunningTime = time.Now().Sub(start).Seconds()
	out.RunStdout = stdout
	out.RunStderr = stderr
	out.Code = code
	out.Signal = signal
	if err == nil {
		out.Status = "Success"
	} else if _, ok := err.(sango.TimeoutError); ok {
		out.Status = "Time limit exceeded"
	} else {
		out.Status = "Runtime error"
	}
}
