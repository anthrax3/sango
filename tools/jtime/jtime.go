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
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return
	}

	var out sango.ExecResult
	out.Command.Run = strings.Join(args, " ")
	var stdout, stderr bytes.Buffer
	msgStdout := sango.MsgpackFilter{Writer: os.Stderr, Tag: "run-stdout"}
	msgStderr := sango.MsgpackFilter{Writer: os.Stderr, Tag: "run-stderr"}
	start := time.Now()
	err, code, signal := sango.Exec(args[0], args[1:], os.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	out.RunningTime = time.Now().Sub(start).Seconds()
	out.Stdout = string(stdout.Bytes())
	out.Stderr = string(stderr.Bytes())
	out.Code = code
	out.Signal = signal

	var usage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_CHILDREN, &usage)
	out.Rusage = sango.Rusage{
		Utime:    float64(usage.Utime.Nano()) / 1000000000.0,
		Stime:    float64(usage.Stime.Nano()) / 1000000000.0,
		Maxrss:   usage.Maxrss,
		Ixrss:    usage.Ixrss,
		Idrss:    usage.Idrss,
		Isrss:    usage.Isrss,
		Minflt:   usage.Minflt,
		Majflt:   usage.Majflt,
		Nswap:    usage.Nswap,
		Inblock:  usage.Inblock,
		Oublock:  usage.Oublock,
		Msgsnd:   usage.Msgsnd,
		Msgrcv:   usage.Msgrcv,
		Nsignals: usage.Nsignals,
		Nvcsw:    usage.Nvcsw,
		Nivcsw:   usage.Nivcsw,
	}

	d := msgpack.NewEncoder(os.Stdout)
	err = d.Encode(out)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Close()
}
