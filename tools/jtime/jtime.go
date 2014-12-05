package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/h2so5/sango/src"
	"github.com/vmihailenco/msgpack"
)

func main() {
	timeout := flag.Duration("t", time.Second*5, "timeout")
	prefix := flag.String("p", "", "prefix")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return
	}

	var out sango.ExecResult
	out.Command = strings.Join(args, " ")
	var stdout, stderr bytes.Buffer
	msgStdout := sango.MsgpackFilter{Writer: os.Stderr, Tag: *prefix + "stdout"}
	msgStderr := sango.MsgpackFilter{Writer: os.Stderr, Tag: *prefix + "stderr"}
	start := time.Now()
	err, code, signal := sango.Exec(args[0], args[1:], os.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), *timeout)
	if err != nil {
		if _, ok := err.(sango.TimeoutError); ok {
			out.Timeout = true
		}
	}
	out.RunningTime = time.Now().Sub(start).Seconds()
	out.Stdout = string(stdout.Bytes())
	out.Stderr = string(stderr.Bytes())
	out.Code = code
	out.Signal = signal

	var usage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, &usage)
	out.Rusage = sango.Rusage{
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

	d := msgpack.NewEncoder(os.Stdout)
	err = d.Encode(out)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Close()
}
