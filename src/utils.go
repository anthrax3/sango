package sango

import (
	"bytes"
	"math/big"
	"math/rand"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/tv42/base58"
)

const BufferSize = 1024

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

type LimitedBuffer struct {
	buf []byte
}

func (b *LimitedBuffer) Write(p []byte) (n int, err error) {
	if len(b.buf)+len(p) > BufferSize {
		l := BufferSize - len(b.buf)
		b.buf = append(b.buf, p[:l]...)
		return l, bytes.ErrTooLarge
	}
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b LimitedBuffer) String() string {
	return string(b.buf)
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "timeout"
}

func Exec(command string, args []string, stdin string, timeout time.Duration) (string, string, error, int, int) {
	cmd := exec.Command(command, args...)
	var stdout, stderr LimitedBuffer
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
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

	return stdout.String(), stderr.String(), err, code, signal
}

func GenerateID() string {
	return string(base58.EncodeBig(nil, big.NewInt(0).Add(big.NewInt(0xc0ffee), big.NewInt(rand.Int63()))))
}
