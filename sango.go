package main

import (
	"flag"
	"log"

	sango "./src"
)

var configFile *string = flag.String("f", "/etc/sango.yml", "Specify config file")
var pull *bool = flag.Bool("p", true, "Pull images")
var build *bool = flag.Bool("b", false, "Build images")
var nocache *bool = flag.Bool("n", false, "Do not use cache on rebuilds")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	conf := sango.LoadConfig(*configFile)
	sango.MakeImageList(conf.ImageDir, *pull, *build, *nocache)
}
