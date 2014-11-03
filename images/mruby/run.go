package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"time"
	"flag"
	"bytes"
	"io"

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
		os.Stdout.Close()
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

	var args []string
	for k, v := range in.Files {
		err := ioutil.WriteFile(k, []byte(v), 0644)
		if err != nil {
			return
		}
		args = append(args, k)
	}

	var stdout, stderr bytes.Buffer
	msgStdout := sango.MsgpackFilter{Writer: os.Stderr, Tag: "run-stdout"}
	msgStderr := sango.MsgpackFilter{Writer: os.Stderr, Tag: "run-stderr"}
	start := time.Now()
	err, code, signal := sango.Exec("/mruby/build/host/bin/mruby", args, in.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	out.RunningTime = time.Now().Sub(start).Seconds()
	out.RunStdout = string(stdout.Bytes())
	out.RunStderr = string(stderr.Bytes())
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
