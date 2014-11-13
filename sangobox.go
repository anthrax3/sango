package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"bitbucket.org/kardianos/osext"
	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/render"
	"github.com/vmihailenco/msgpack"

	"github.com/h2so5/sango/src"
)

var sangoPath string
var configFile *string = flag.String("f", "/etc/sango.yml", "Specify config file")
var cmdCacheSeconds = 60 * 60

type Sango struct {
	*martini.ClassicMartini
	conf  sango.Config
	db    redis.Conn
	imgch chan sango.ImageList
	reqch chan int
}

func NewSango(conf sango.Config) *Sango {
	m := martini.Classic()
	m.Use(gzip.All())
	m.Use(martini.Static(filepath.Join(sangoPath, "public")))
	m.Use(render.Renderer(render.Options{
		Layout:     "layout",
		Extensions: []string{".html"},
	}))

	eaddr := os.Getenv("REDIS_PORT_6379_TCP_ADDR")
	eport := os.Getenv("REDIS_PORT_6379_TCP_PORT")

	addr := ":6379"
	if len(eaddr) > 0 && len(eport) > 0 {
		addr = eaddr + ":" + eport
	}

	db, err := redis.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	s := &Sango{
		ClassicMartini: m,
		conf:           conf,
		db:             db,
		imgch:          make(chan sango.ImageList),
		reqch:          make(chan int, conf.ExecLimit),
	}

	imgdch := make(chan sango.ImageList)
	go func() {
		tick := time.Tick(1 * time.Hour)
		for {
			images, err := sango.MakeImageList(s.conf.ImageDir, false)
			if err != nil {
				imgdch <- images
			}
			<-tick
		}
	}()

	go func() {
		var images sango.ImageList
		data, err := redis.Bytes(s.db.Do("GET", "images"))
		if err == nil {
			err = msgpack.Unmarshal(data, &images)
		}
		for {
			select {
			case i := <-imgdch:
				images = i
				data, err := msgpack.Marshal(images)
				if err != nil {
					log.Print(err)
				} else {
					_, err := s.db.Do("SET", "images", data)
					if err != nil {
						log.Print(err)
					}
				}
			case s.imgch <- images:
			}
		}
	}()

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
		r.Post("/cmd", s.apiCmd)
		r.Get("/run/stream", s.apiRunStreaming)
		r.Get("/log/:id", s.apiLog)
	})

	m.Get("/", s.index)
	m.Get("/:id", s.log)
	m.Get("/template/:env", s.template)
	m.Get("/hello/:env", s.hello)

	return s
}

func (s *Sango) getImageList() sango.ImageList {
	data, err := redis.Bytes(s.db.Do("GET", "images"))

	var list sango.ImageList
	if err == nil {
		err = msgpack.Unmarshal(data, &list)
		log.Print(err)
	}
	return list
}

func (s *Sango) images() sango.ImageList {
	return <-s.imgch
}

func (s *Sango) index(r render.Render) {
	r.HTML(200, "index", map[string]interface{}{
		"ga":     s.conf.GoogleAnalytics,
		"images": s.imageArray(),
	})
}

func (s *Sango) log(r render.Render, params martini.Params) {
	id := params["id"]

	n, err := redis.Bool(s.db.Do("EXISTS", "log/"+id))
	if err != nil || !n {
		r.Redirect("/")
		return
	}
	r.HTML(200, "index", map[string]interface{}{
		"ga":     s.conf.GoogleAnalytics,
		"logid":  id,
		"images": s.imageArray(),
	})
}

func (s *Sango) imageArray() []sango.Image {
	images := s.images()
	l := make(sango.ImageArray, 0, len(images))
	for _, v := range images {
		l = append(l, v)
	}
	sort.Sort(l)
	return l
}

func (s *Sango) apiImageList(r render.Render) {
	r.JSON(200, s.imageArray())
}

func (s *Sango) run(req io.Reader, msgch chan<- *sango.Message) (ExecResponse, int, error) {
	reader := io.LimitReader(req, s.conf.UploadLimit)
	d := json.NewDecoder(reader)
	var ereq ExecRequest
	err := d.Decode(&ereq)
	if err != nil {
		log.Print(err)
		if reader.(*io.LimitedReader).N <= 0 {
			return ExecResponse{}, 413, errors.New("Too large input")
		} else {
			return ExecResponse{}, 400, errors.New("Bad request")
		}
	}
	if len(ereq.Input.Files) == 0 {
		return ExecResponse{}, 400, errors.New("No input files")
	}
	img, ok := s.images()[ereq.Environment]
	if !ok {
		return ExecResponse{}, 501, errors.New("No such environment")
	}
	s.reqch <- 0
	defer func() { <-s.reqch }()

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
			_, err := s.db.Do("SET", "log/"+eres.ID, data)
			if err != nil {
				log.Print(err)
			}
		}
	}
	return eres, 200, nil
}

func (s *Sango) apiRun(r render.Render, res http.ResponseWriter, req *http.Request) {
	eres, code, err := s.run(req.Body, nil)
	if err != nil {
		r.JSON(code, map[string]string{"error": err.Error()})
	} else {
		r.JSON(code, eres)
	}
}

func (s *Sango) getCmd(req ExecRequest) (sango.CommandLine, int, error) {
	var c sango.CommandLine
	data, err := msgpack.Marshal(req)
	if err != nil {
		return c, 500, errors.New("Internal error")
	}

	id := md5.Sum(data)
	data, err = redis.Bytes(s.db.Do("GET", "cache/cmd/"+string(id[:])))
	if err == nil {
		err := msgpack.Unmarshal(data, &c)
		if err != nil {
			log.Print(err)
		} else {
			return c, 200, nil
		}
	}

	img, ok := s.images()[req.Environment]
	if !ok {
		return c, 501, errors.New("No such environment")
	}

	cmd, err := img.GetCommand(req.Input)
	if err != nil {
		return c, 500, errors.New("Internal error")
	}

	c = cmd

	data, err = msgpack.Marshal(c)
	if err != nil {
		log.Print(err)
	} else {
		_, err := s.db.Do("SETEX", "cache/cmd/"+string(id[:]), cmdCacheSeconds, data)
		if err != nil {
			log.Print(err)
		}
	}

	return c, 200, nil
}

func (s *Sango) apiCmd(r render.Render, res http.ResponseWriter, req *http.Request) {
	reader := io.LimitReader(req.Body, s.conf.UploadLimit)
	d := json.NewDecoder(reader)
	var ereq ExecRequest
	err := d.Decode(&ereq)

	if err != nil {
		r.JSON(400, map[string]string{"error": "Bad request"})
		return
	}

	cmd, code, err := s.getCmd(ereq)

	if err != nil {
		r.JSON(code, map[string]string{"error": err.Error()})
	} else {
		r.JSON(200, cmd)
	}
}

func (s *Sango) apiRunStreaming(res http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
	if err != nil {
		log.Print(err)
		return
	}
	defer ws.Close()

	_, r, err := ws.NextReader()
	if err != nil {
		log.Print(err)
		return
	}

	msgch := make(chan *sango.Message)
	go func() {
		for {
			msg := <-msgch
			if msg == nil {
				return
			}
			ws.WriteJSON(msg)
		}
	}()

	eres, _, err := s.run(r, msgch)
	if err != nil {
		log.Print(err)
		return
	}

	if err != nil {
		ws.WriteJSON(map[string]string{"error": err.Error()})
	} else {
		ws.WriteJSON(map[string]interface{}{"tag": "result", "data": eres})
	}
}

func (s *Sango) apiLog(r render.Render, params martini.Params) {
	data, err := redis.Bytes(s.db.Do("GET", "log/"+params["id"]))
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

func (s *Sango) template(res http.ResponseWriter, params martini.Params) {
	env := params["env"]
	img, ok := s.images()[env]
	res.Header()["Content-Type"] = []string{"text/plain"}
	if !ok {
		res.WriteHeader(404)
		return
	}
	res.WriteHeader(200)
	res.Write([]byte(img.Template))
}

func (s *Sango) hello(res http.ResponseWriter, params martini.Params) {
	env := params["env"]
	img, ok := s.images()[env]
	res.Header()["Content-Type"] = []string{"text/plain"}
	if !ok {
		res.WriteHeader(404)
		return
	}
	res.WriteHeader(200)
	res.Write([]byte(img.HelloWorld))
}

func (s *Sango) Close() {
	s.db.Close()
}

type ExecRequest struct {
	Environment string      `json:"environment"`
	Volatile    bool        `json:"volatile"`
	Input       sango.Input `json:"input"`
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

	conf := sango.LoadConfig(*configFile)
	s := NewSango(conf)
	defer s.Close()
	log.Printf("listening on :%d\n", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), s))
}
