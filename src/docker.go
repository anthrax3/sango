package sango

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"
)

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
	cmd := exec.Command("docker", "run", "--rm", "-i", "--net='none'", i.dockerImageName(), "cat", "/sango/version")
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	} else {
		return stdout.String(), nil
	}
}

func (i Image) Exec(in Input) (Output, error) {
	data, err := msgpack.Marshal(in)
	if err != nil {
		return Output{}, err
	}
	id := GenerateID()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", "run", "--rm", "-i", "--name", id, "--net='none'", i.dockerImageName(), "./run")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Second * 5):
		stopcmd := exec.Command("docker", "stop", "--time=0", id)
		stopcmd.Run()
		err = <-ch
	case err = <-ch:
	}

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

	c := "--no-cache="
	if nocache {
		c += "true"
	} else {
		c += "false"
	}
	cmd = exec.Command("docker", "build", c, "-t", image, ".")
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
