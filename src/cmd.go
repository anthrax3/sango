package sango

import "strings"

type CmdCommand struct {
	Input
	H map[string]func(Input) (cmd [][]string)
}

func (c CmdCommand) Invoke() interface{} {
	m := make(map[string][]string)
	for stage, h := range c.H {
		for _, cmd := range h(c.Input) {
			m[stage] = append(m[stage], strings.Join(cmd, " "))
		}
	}
	return m
}
