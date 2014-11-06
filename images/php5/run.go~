package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"time"

	sango "../../src"

	"github.com/vmihailenco/msgpack"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

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

	var args []string
	for k, v := range in.Files {
		err := ioutil.WriteFile(k, []byte(v), 0644)
		if err != nil {
			return
		}
		args = append(args, k)
	}

	start := time.Now()
	stdout, stderr, err, code, signal := sango.Exec("/mruby/build/host/bin/mruby", args, in.Stdin, 5*time.Second)
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
