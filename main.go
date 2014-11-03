package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"bitbucket.org/kardianos/osext"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vmihailenco/msgpack"
	"gopkg.in/yaml.v2"

	sango "./src"
)

var sangoPath string

var forceBuild *bool = flag.Bool("b", false, "Force to rebuild all docker images on startup")
var configFile *string = flag.String("f", "/etc/sango.yml", "Specify config file")
var noCache *bool = flag.Bool("nocache", false, "Do not use cache on rebuilds")
var noRun *bool = flag.Bool("norun", false, "Do not run server")

type Sango struct {
	*martini.ClassicMartini
	conf   Config
	db     *leveldb.DB
	images sango.ImageList
	reqch  chan int
}

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

func NewSango(conf Config) *Sango {
	m := martini.Classic()
	m.Use(gzip.All())
	m.Use(martini.Static(filepath.Join(sangoPath, "public")))
	m.Use(render.Renderer(render.Options{
		Layout:     "layout",
		Extensions: []string{".html"},
	}))

	db, err := leveldb.OpenFile(conf.Database, nil)
	if err != nil {
		log.Fatal(err)
	}

	sango.CleanImages()
	images := sango.MakeImageList(conf.ImageDir, *forceBuild, *noCache)

	s := &Sango{
		ClassicMartini: m,
		conf:           conf,
		db:             db,
		images:         images,
		reqch:          make(chan int, conf.ExecLimit),
	}

	ch := time.Tick(conf.CleanInterval)
	go func() {
		for {
			<-ch
			log.Print("cleaning images...")
			sango.CleanImages()
		}
	}()

	m.Group("/api", func(r martini.Router) {
		r.Get("/list", s.apiImageList)
		r.Post("/run", s.apiRun)
		r.Get("/log/:id", s.apiLog)
	})

	m.Get("/", s.index)
	m.Get("/:id", s.log)

	return s
}

func (s *Sango) index(r render.Render) {
	r.HTML(200, "index", map[string]interface{}{"images": s.imageArray()})
}

func (s *Sango) log(r render.Render, params martini.Params) {
	id := params["id"]
	_, err := s.db.Get([]byte(id), nil)
	if err != nil {
		r.Redirect("/")
		return
	}
	r.HTML(200, "index", map[string]interface{}{"logid": id, "images": s.imageArray()})
}

func (s *Sango) imageArray() []sango.Image {
	l := make(sango.ImageArray, 0, len(s.images))
	for _, v := range s.images {
		l = append(l, v)
	}
	sort.Sort(l)
	return l
}

func (s *Sango) apiImageList(r render.Render) {
	r.JSON(200, s.imageArray())
}

func (s *Sango) apiRun(r render.Render, res http.ResponseWriter, req *http.Request) {
	reader := io.LimitReader(req.Body, s.conf.UploadLimit)
	d := json.NewDecoder(reader)
	var ereq ExecRequest
	err := d.Decode(&ereq)
	if err != nil {
		log.Print(err)
		if reader.(*io.LimitedReader).N <= 0 {
			r.JSON(413, map[string]string{"error": "Too large input"})
		} else {
			r.JSON(400, map[string]string{"error": "Bad request"})
		}
		return
	}
	if len(ereq.Input.Files) == 0 {
		r.JSON(400, map[string]string{"error": "No input files"})
		return
	}
	img, ok := s.images[ereq.Environment]
	if !ok {
		r.JSON(501, map[string]string{"error": "No such environment"})
		return
	} else {
		s.reqch <- 0
		defer func() { <-s.reqch }()

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(200)

		var msgch chan *sango.Message
		if ereq.Streaming {
			msgch = make(chan *sango.Message)
			go func() {
				for {
					m := <- msgch
					if m == nil {
						return
					}
					log.Print(m)
					data, _ := json.Marshal(m)
					res.Write(data)
					res.(http.Flusher).Flush()
				}
			}()
		}

		out, err := img.Exec(ereq.Input, msgch)
		if err != nil {
			log.Print(err)
		}
		eres := ExecResponse{
			Environment: img,
			Input:       ereq.Input,
			Output:      out,
			Date:        time.Now(),
		}
		if !ereq.Volatile {
			eres.ID = sango.GenerateID()
			data, err := msgpack.Marshal(eres)
			if err != nil {
				log.Print(err)
			} else {
				err := s.db.Put([]byte(eres.ID), data, nil)
				if err != nil {
					log.Print(err)
				}
			}
		}
		data, _ := json.Marshal(eres)
		res.Write(data)
	}
}

func (s *Sango) apiLog(r render.Render, params martini.Params) {
	data, err := s.db.Get([]byte(params["id"]), nil)
	if err != nil {
		log.Print(err)
		r.JSON(404, map[string]string{"error": "Not found"})
		return
	}
	var res ExecResponse
	err = msgpack.Unmarshal(data, &res)
	if err != nil {
		log.Print(err)
		r.JSON(500, map[string]string{"error": "Internal error"})
		return
	}
	r.JSON(200, res)
}

func (s *Sango) Close() {
	s.db.Close()
}

type ExecRequest struct {
	Environment string      `json:"environment"`
	Volatile    bool        `json:"volatile"`
	Input       sango.Input `json:"input"`
	Streaming   bool        `json:"streaming"`
}

type ExecResponse struct {
	ID          string       `json:"id,omitempty"`
	Environment sango.Image  `json:"environment"`
	Input       sango.Input  `json:"input"`
	Output      sango.Output `json:"output"`
	Date        time.Time    `json:"date"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	rand.Seed(time.Now().Unix())
	runtime.GOMAXPROCS(runtime.NumCPU())

	path, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(path)
	if err != nil {
		log.Fatal(err)
	}
	sangoPath = path

	conf := LoadConfig(*configFile)
	if !*noRun {
		s := NewSango(conf)
		defer s.Close()
		log.Printf("listening on :%d\n", conf.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), s))
	} else {
		sango.MakeImageList(conf.ImageDir, *forceBuild, *noCache)
	}
}
