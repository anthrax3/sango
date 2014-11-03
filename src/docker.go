package sango

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"
	"io"

	"github.com/h2so5/docker/api/client"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"
)

const dockerAddr = "/var/run/docker.sock"

type Image struct {
	ID       string `yaml:"id"        json:"id"`
	Name     string `yaml:"name"      json:"name"`
	Language string `yaml:"language"  json:"language"`
	Version  string `yaml:"-"         json:"version"`
}

func (i Image) dockerImageName() string {
	return "sango-" + i.ID
}

func (i Image) GetVersion() (string, error) {
	var stdout bytes.Buffer
	c := client.NewDockerCli(nil, &stdout, nil, "unix", dockerAddr, nil)
	if c == nil {
		return "", errors.New("failed to create docker client")
	}
	err := c.CmdRun("-i", "--net=none", i.dockerImageName(), "./run", "-v")
	if err != nil {
		return "", err
	} else {
		return stdout.String(), nil
	}
}

func (i Image) Exec(in Input, msgch chan<- *Message) (Output, error) {
	data, err := msgpack.Marshal(in)
	if err != nil {
		return Output{}, err
	}
	id := GenerateID()

	var stdout bytes.Buffer
	r, w := io.Pipe()
	c := client.NewDockerCli(NewCloserReader(data), &stdout, w, "unix", dockerAddr, nil)
	if c == nil {
		return Output{}, errors.New("failed to create docker client")
	}

	ch := make(chan error, 1)
	go func() {
		ch <- c.CmdRun("-i", "--name", id, "--net=none", i.dockerImageName(), "nice", "-n", "20", "./run")
	}()

	go func() {
		d := msgpack.NewDecoder(r)
		for {
			var m Message
			err := d.Decode(&m)
			if err != nil {
				if msgch != nil {
					close(msgch)
				}
				return
			}
			if msgch != nil {
				msgch <- &m
			}
		}
	}()

	select {
	case <-time.After(time.Second * 8):
		c.CmdStop("--time=0", id)
		err = <-ch
	case err = <-ch:
	}

	r.Close()
	w.Close()

	var out Output
	if err != nil {
		out.Status = "Internal error"
	} else {
		err = msgpack.Unmarshal(stdout.Bytes(), &out)
		if err != nil {
			return Output{}, err
		}
	}

	return out, nil
}

func (i Image) exists() bool {
	var stdout bytes.Buffer
	c := client.NewDockerCli(nil, &stdout, nil, "unix", dockerAddr, nil)
	if c == nil {
		return false
	}
	err := c.CmdInspect(i.dockerImageName())
	return err == nil
}

func buildImage(dir, image string, nocache bool) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", "run", "run.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	nc := "--no-cache="
	if nocache {
		nc += "true"
	} else {
		nc += "false"
	}

	c := client.NewDockerCli(nil, os.Stdout, os.Stderr, "unix", dockerAddr, nil)
	if c == nil {
		return errors.New("failed to create docker client")
	}

	err = c.CmdBuild(nc, "-t", image, ".")
	if err != nil {
		return err
	}

	err = os.Remove("run")
	if err != nil {
		return err
	}

	err = os.Chdir(pwd)
	if err != nil {
		return err
	}

	return nil
}

func CleanImages() error {
	var stdout bytes.Buffer
	c := client.NewDockerCli(nil, &stdout, os.Stderr, "unix", dockerAddr, nil)
	if c == nil {
		return errors.New("failed to create docker client")
	}

	err := c.CmdPs("-a", "-q")
	if err != nil {
		return err
	}
	ps := strings.Split(string(stdout.Bytes()), "\n")
	if (len(ps) > 1) {
		err = c.CmdRm(ps[:len(ps)-1]...)
		if err != nil {
			return err
		}
	}

	return nil
}

type ImageList map[string]Image

func MakeImageList(langpath string, build, nocache bool) ImageList {
	l := make(ImageList)

	info, err := os.Stat(langpath)
	if err != nil {
		log.Print(err)
		return nil
	}

	if !info.IsDir() {
		log.Print("%s is not a directory", langpath)
		return nil
	}

	files, err := ioutil.ReadDir(langpath)
	if err != nil {
		log.Print(err)
		return nil
	}

	for _, f := range files {
		d := filepath.Join(langpath, f.Name())
		c := filepath.Join(d, "config.yml")
		data, err := ioutil.ReadFile(c)
		if err == nil {
			var img Image
			err := yaml.Unmarshal(data, &img)
			if err != nil {
				log.Print(c, err)
			} else {
				if build {
					log.Printf("Found config: %s [%s]", img.ID, img.dockerImageName())
					log.Printf("Building image...")
					err = buildImage(d, img.dockerImageName(), nocache)
				} else {
					if !img.exists() {
						log.Printf("Image not found: %s", img.dockerImageName())
						continue
					}
				}
				if err != nil {
					log.Printf("Filed to build image: %v", err)
				} else {
					ver, err := img.GetVersion()
					if err != nil {
						log.Printf("Filed to get version: %v", err)
					} else {
						img.Version = ver
						log.Printf("Get version: %s (%s)", img.Language, img.Version)
						l[img.ID] = img
					}
				}
			}
		}
	}

	return l
}

type ImageArray []Image

func (a ImageArray) Len() int {
	return len(a)
}

func (a ImageArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ImageArray) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}
