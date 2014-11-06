package sango

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port            uint16        `yaml:"port"`
	Database        string        `yaml:"database"`
	ImageDir        string        `yaml:"image_dir"`
	UploadLimit     int64         `yaml:"upload_limit"`
	ExecLimit       int           `yaml:"exec_limit"`
	CleanInterval   time.Duration `yaml:"clean_interval"`
	GoogleAnalytics string        `yaml:"google_analytics"`
}

func defaultConfig() Config {
	return Config{
		Port:            3000,
		Database:        "./sango.leveldb",
		ImageDir:        "./images",
		UploadLimit:     20480,
		CleanInterval:   time.Minute,
		ExecLimit:       5,
		GoogleAnalytics: "",
	}
}

func LoadConfig(path string) Config {
	c := defaultConfig()
	data, err := ioutil.ReadFile(path)
	if err == nil {
		err := yaml.Unmarshal(data, &c)
		if err != nil {
			log.Print(err)
		}
	} else {
		log.Print(err)
	}
	return c
}
