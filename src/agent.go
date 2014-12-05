package sango

import (
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/vmihailenco/msgpack"
)

const LimitedWriterSize = 1024 * 10

type MsgpackFilter struct {
	Writer io.Writer
	Tag    string
}

func (j *MsgpackFilter) Write(p []byte) (n int, err error) {
	v := Message{
		Tag:  j.Tag,
		Data: string(p),
	}
	data, err := msgpack.Marshal(v)
	if err != nil {
		return 0, err
	}
	_, err = j.Writer.Write(data)
	return len(p), err
}

type Message struct {
	Tag  string `msgpack:"t" json:"tag"`
	Data string `msgpack:"d" json:"data"`
}

func Exec(command string, args []string, stdin io.Reader, rstdout, rstderr io.Writer, timeout time.Duration) (error, int, int) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = stdin
	cmd.Stdout = &LimitedWriter{W: rstdout, N: LimitedWriterSize}
	cmd.Stderr = &LimitedWriter{W: rstderr, N: LimitedWriterSize}
	cmd.Start()

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	var timech <-chan time.Time
	if timeout != 0 {
		timech = time.After(timeout)
	}

	var err error
	var timeouterr bool
	select {
	case <-timech:
		cmd.Process.Kill()
		err = <-ch
		timeouterr = true
	case err = <-ch:
	}

	var code, signal int
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				code = status.ExitStatus()
				signal = int(status.StopSignal())
			}
		}
	}
	if timeouterr {
		err = TimeoutError{}
	}

	return err, code, signal
}

type LimitedWriter struct {
	W io.Writer
	N int64
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.ErrClosedPipe
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}

type Option struct {
	Title      string        `yaml:"title"      json:"title"`
	Type       string        `yaml:"type"       json:"type"`
	Default    interface{}   `yaml:"default"    json:"default"`
	Candidates []interface{} `yaml:"candidates" json:"candidates,omitempty"`
}

type Input struct {
	Files   map[string]string      `json:"files"`
	Stdin   string                 `json:"stdin"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type Output struct {
	Results     map[string]ExecResult `json:"results"`
	MixedOutput []Message             `json:"mixed-output"`
	Status      string                `json:"status"`
}

type ExecResult struct {
	Stdout      string  `json:"stdout"`
	Stderr      string  `json:"stderr"`
	Rusage      Rusage  `json:"rusage"`
	Command     string  `json:"command"`
	Code        int     `json:"code"`
	Signal      int     `json:"signal"`
	RunningTime float64 `json:"running-time"`
	Timeout     bool    `json:"timeout"`
}

type Rusage struct {
	Utime   float64 `json:"utime"`
	Stime   float64 `json:"stime"`
	Maxrss  int64   `json:"maxrss"`
	Minflt  int64   `json:"minflt"`
	Majflt  int64   `json:"majflt"`
	Inblock int64   `json:"inblock"`
	Oublock int64   `json:"oublick"`
	Nvcsw   int64   `json:"nvcsw"`
	Nivcsw  int64   `json:"nivcsw"`
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "timeout"
}
