package sango

import (
	"bytes"
	"io"
	"math/big"
	"math/rand"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/tv42/base58"
	"github.com/vmihailenco/msgpack"
)

const LimitedWriterSize = 1024 * 500

type Input struct {
	Files map[string]string `json:"files"`
	Stdin string            `json:"stdin"`
}

type Output struct {
	BuildStdout string  `json:"build-stdout"`
	BuildStderr string  `json:"build-stderr"`
	RunStdout   string  `json:"run-stdout"`
	RunStderr   string  `json:"run-stderr"`
	Code        int     `json:"code"`
	Signal      int     `json:"signal"`
	Status      string  `json:"status"`
	RunningTime float64 `json:"running-time"`
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "timeout"
}

type CloserReader struct {
	*bytes.Reader
}

func NewCloserReader(b []byte) *CloserReader {
	return &CloserReader{
		Reader: bytes.NewReader(b),
	}
}

type Message struct {
	Tag  string `msgpack:"t" json:"tag"`
	Data string `msgpack:"d" json:"data"`
}

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

func (c CloserReader) Close() error {
	return nil
}

func Exec(command string, args []string, stdin string, rstdout, rstderr io.Writer, timeout time.Duration) (error, int, int) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = strings.NewReader(stdin)
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

func GenerateID() string {
	return string(base58.EncodeBig(nil, big.NewInt(0).Add(big.NewInt(0xc0ffee), big.NewInt(rand.Int63()))))
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
