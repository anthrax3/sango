package sango

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/tv42/base58"
	"github.com/vmihailenco/msgpack"
)

const dockerAddr = "/var/run/docker.sock"
const imagePrefix = "sango/"
const hubImageListEndpoint = "https://index.docker.io/v1/search?q=" + imagePrefix

type Image struct {
	ID         string            `yaml:"id"         json:"id"`
	Name       string            `yaml:"name"       json:"name"`
	Language   string            `yaml:"language"   json:"language"`
	Options    map[string]Option `yaml:"options"    json:"options,omitempty"`
	Version    string            `yaml:"-"          json:"version"`
	Protocol   int               `yaml:"-"          json:"-"`
	Template   string            `yaml:"-"          json:"-"`
	HelloWorld string            `yaml:"-"          json:"-"`
	Extensions []string          `yaml:"extensions" json:"extensions"`
	AceMode    string            `yaml:"acemode"    json:"-"`
}

func (i Image) dockerImageName() string {
	return imagePrefix + i.ID
}

func (i *Image) GetInfo() error {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "run", "-i", "--net=none", i.dockerImageName(), "agent", "version")
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		return err
	} else {
		err := msgpack.Unmarshal(stdout.Bytes(), i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Image) GetCommand(in Input) (map[string]string, error) {
	var c map[string]string
	data, err := msgpack.Marshal(in)
	if err != nil {
		return c, err
	}

	var stdout bytes.Buffer
	cmd := exec.Command("docker", "run", "-i", "--net=none", i.dockerImageName(), "agent", "cmd")
	cmd.Stdin = bytes.NewBuffer(data)
	cmd.Stdout = &stdout
	err = cmd.Run()

	if err != nil {
		return c, err
	} else {
		err := msgpack.Unmarshal(stdout.Bytes(), &c)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func GenerateID() string {
	return string(base58.EncodeBig(nil, big.NewInt(0).Add(big.NewInt(0xc0ffee), big.NewInt(rand.Int63()))))
}

func (i Image) Exec(in Input, msgch chan<- *Message) (Output, error) {
	data, err := msgpack.Marshal(in)
	if err != nil {
		return Output{}, err
	}
	id := GenerateID()

	if in.Options == nil {
		in.Options = make(map[string]interface{})
	}

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
	cmd := exec.Command("docker", "run", "-i", "--name", id, "--net=none", i.dockerImageName(), "agent", "run")
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = &stdout
	cmd.Stderr = w
	cmd.Start()

	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	out := Output{
		MixedOutput: make([]Message, 0),
	}

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
			return Output{}, err
		}
	}

	return out, nil
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

func pullImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type HubImageList struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func getHubImages() ([]string, error) {
	resp, err := http.Get(hubImageListEndpoint)
	if err != nil {
		return nil, err
	}
	var h HubImageList
	j := json.NewDecoder(resp.Body)
	err = j.Decode(&h)
	if err != nil {
		return nil, err
	}
	var l []string
	for _, n := range h.Results {
		name := strings.Replace(n.Name, imagePrefix, "", 1)
		if !strings.HasPrefix(name, "_") {
			l = append(l, name)
		}
	}
	return l, nil
}

var idRegexp = regexp.MustCompile("^" + regexp.QuoteMeta(imagePrefix) + "[^_].+? ")

func images() ([]string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "images")
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	var i []string
	for _, l := range strings.Split(stdout.String(), "\n") {
		id := strings.Trim(strings.Replace(string(idRegexp.Find([]byte(l))), imagePrefix, "", -1), " ")
		if len(id) > 0 {
			i = append(i, id)
		}
	}
	return i, nil
}

type ImageList map[string]Image

func MakeImageList(langpath string, pull bool) (ImageList, error) {
	l := make(ImageList)

	var imgs []string
	var err error
	if pull {
		imgs, err = getHubImages()
		if err != nil {
			log.Print("Filed to get image list from the hub: %v", err)
		} else {
			log.Print("Get image list from the hub")
		}
	}

	if !pull || err != nil {
		imgs, err = images()
	}

	if err != nil {
		return nil, err
	}

	for _, i := range imgs {
		img := Image{ID: i}
		if pull {
			err := pullImage(img.dockerImageName())
			if err != nil {
				log.Print(err)
			}
		}
		err := img.GetInfo()
		if err != nil {
			log.Printf("Filed to get version: %v", err)
		} else {
			if img.Protocol != ProtocolVersion {
				log.Printf("Protocol version mismatch: %s (%s) %d", img.Language, img.Version, img.Protocol)
			} else {
				log.Printf("Loaded: %s (%s)", img.Language, img.Version)
				l[img.ID] = img
			}
		}
	}

	return l, nil
}

type ImageArray []Image

func (a ImageArray) Len() int {
	return len(a)
}

func (a ImageArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ImageArray) Less(i, j int) bool {
	return a[i].Language < a[j].Language
}
