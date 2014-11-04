package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tv42/base58"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"

	"./agent"
)

const dockerAddr = "/var/run/docker.sock"
const imagePrefix = "sango/"

type Image struct {
	ID       string                  `yaml:"id"       json:"id"`
	Name     string                  `yaml:"name"     json:"name"`
	Language string                  `yaml:"language" json:"language"`
	Options  map[string]agent.Option `yaml:"options"  json:"options,omitempty"`
	Version  string                  `yaml:"-"        json:"version"`
	Template string                  `yaml:"-"        json:"-"`
	AceMode  string                  `yaml:"acemode"  json:"-"`
}

func (i Image) dockerImageName() string {
	return imagePrefix + i.ID
}

func (i Image) GetVersion() (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "run", "-i", "--net=none", i.dockerImageName(), "./run", "-v")
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	} else {
		return stdout.String(), nil
	}
}

func GenerateID() string {
	return string(base58.EncodeBig(nil, big.NewInt(0).Add(big.NewInt(0xc0ffee), big.NewInt(rand.Int63()))))
}

func (i Image) Exec(in agent.Input, msgch chan<- *agent.Message) (agent.Output, error) {
	data, err := msgpack.Marshal(in)
	if err != nil {
		return agent.Output{}, err
	}
	id := GenerateID()

	for k, v := range i.Options {
		if _, ok := in.Options[k]; !ok {
			in.Options[k] = v.Default
		}
		o := in.Options[k]
		in.Options[k] = v.Default
		switch v.Type {
		case "list":
			if s, ok := o.(string); ok {
				for _, i := range v.Candidates {
					if s == i.(string) {
						in.Options[k] = s
					}
				}
			}
		case "bool":
			if b, ok := o.(bool); ok {
				in.Options[k] = b
			}
		}
	}

	var stdout bytes.Buffer
	r, w := io.Pipe()
	cmd := exec.Command("docker", "run", "-i", "--name", id, "--net=none", i.dockerImageName(), "./run")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = &stdout
	cmd.Stderr = w
	cmd.Start()

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	out := agent.Output{
		MixedOutput: make([]agent.Message, 0),
	}

	go func() {
		d := msgpack.NewDecoder(r)
		for {
			var m agent.Message
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
			out.MixedOutput = append(out.MixedOutput, m)
		}
	}()

	select {
	case <-time.After(time.Second * 8):
		stopcmd := exec.Command("docker", "stop", "--time=0", id)
		stopcmd.Run()
		err = <-ch
	case err = <-ch:
	}

	r.Close()
	w.Close()

	if err != nil {
		out.Status = "Internal error"
	} else {
		err = msgpack.Unmarshal(stdout.Bytes(), &out)
		if err != nil {
			return agent.Output{}, err
		}
	}

	return out, nil
}

func (i Image) exists() bool {
	cmd := exec.Command("docker", "inspect", i.dockerImageName())
	err := cmd.Run()
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

	cmd = exec.Command("docker", "build", nc, "-t", image, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
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
	cmd := exec.Command("docker", "ps", "-a", "-q")
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	ps := strings.Split(string(stdout.Bytes()), "\n")
	if len(ps) > 1 {
		cmd := exec.Command("docker", append([]string{"rm"}, ps[:len(ps)-1]...)...)
		err = cmd.Run()
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

	w, err := os.Getwd()
	if err != nil {
		log.Print(err)
		return nil
	}

	for _, f := range files {
		err = os.Chdir(w)
		if err != nil {
			log.Print(w)
			continue
		}
		d := filepath.Join(langpath, f.Name())
		c := filepath.Join(d, "config.yml")
		data, err := ioutil.ReadFile(c)
		if err != nil {
			log.Print(w)
			continue
		}
		var img Image
		err = yaml.Unmarshal(data, &img)
		if err != nil {
			log.Print(c, err)
		} else {
			t := filepath.Join(d, "template.txt")
			data, _ := ioutil.ReadFile(t)
			img.Template = string(data)
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
