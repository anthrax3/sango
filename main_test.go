package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"./src"
)

func TestAPI(t *testing.T) {
	conf := LoadConfig("")
	s := NewSango(conf)
	defer s.Close()

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), s)
		if err != nil {
			t.Fatal(err)
		}
	}()

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/api/list", conf.Port))
	if err != nil {
		t.Fatal(err)
	}

	var images []sango.Image

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(data, &images)
	if err != nil {
		t.Fatal(err)
	}

	for _, img := range images {
		res, err := http.Get(fmt.Sprintf("http://localhost:%d/hello/%s", conf.Port, img.ID))
		if err != nil {
			t.Fatal(err)
		}
		req := ExecRequest{
			Environment: img.ID,
			Volatile:    true,
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		req.Input.Files = make(map[string]string)
		req.Input.Files["main."+img.Extensions[0]] = string(data)
		js, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		res, err = http.Post(fmt.Sprintf("http://localhost:%d/api/run", conf.Port), "application/json", bytes.NewReader(js))
		if err != nil {
			t.Fatal(err)
		}

		var result ExecResponse
		data, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(data, &result)
		if err != nil {
			t.Fatal(err)
		}

		if result.Output.RunStdout != "Hello World" {
			t.Fatalf("%s: The program should return 'Hello World'; not '%s'", img.ID, result.Output.RunStdout)
		}
	}
}
