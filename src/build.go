package sango

import (
	"bytes"
	"io"
	"os"
	"time"
)

type BuildCommand struct {
	Input
	H func(Input) (cmd [][]string)
}

func (c BuildCommand) Invoke() interface{} {
	var stage Stage
	var stdout, stderr bytes.Buffer
	for _, cmd := range c.H(c.Input) {
		if len(cmd) == 0 {
			continue
		}
		msgStdout := MsgpackFilter{Writer: os.Stderr, Tag: "stdout", Stage: "build"}
		msgStderr := MsgpackFilter{Writer: os.Stderr, Tag: "stderr", Stage: "build"}
		err, code, signal := Exec(cmd[0], cmd[1:], c.Input.Stdin, io.MultiWriter(&msgStdout, &stdout), io.MultiWriter(&msgStderr, &stderr), 5*time.Second)
		stage.Stdout = string(stdout.Bytes())
		stage.Stderr = string(stderr.Bytes())
		stage.Code = code
		stage.Signal = signal
		if err != nil {
			if _, ok := err.(TimeoutError); ok {
				stage.Status = "Time limit exceeded"
			} else {
				stage.Status = "Error"
			}
			break
		} else {
			stage.Status = "OK"
		}
	}
	return stage
}
