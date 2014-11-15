package sango

import (
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type VersionCommand struct {
	H func() string
}

func (c VersionCommand) Invoke() interface{} {
	var img Image
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil
	}

	err = yaml.Unmarshal(data, &img)
	if err != nil {
		return nil
	}

	data, _ = ioutil.ReadFile("template.txt")
	img.Template = string(data)

	ver := strings.Trim(c.H(), "\r\n ")
	img.Version = ver
	return img
}
