package sango

import (
	"bytes"
	"log"
	"time"
)

type TestCommand struct {
	Input
	H []func() (cmd [][]string)
}

func (c TestCommand) Invoke() interface{} {
	for _, h := range c.H {
		for _, cmd := range h() {
			var stdout, stderr bytes.Buffer
			if len(cmd) == 0 {
				continue
			}
			err, _, _ := Exec(cmd[0], cmd[1:], "", &stdout, &stderr, 5*time.Second)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}
