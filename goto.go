package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var routes []route

type route struct {
	Path   string
	Target string
	Code   int
}

func main() {
	raw, _ := ioutil.ReadFile("redirect.json")
	err := json.Unmarshal(raw, &routes)

	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", redirectHandler)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("Serve: ", err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.RequestURI()

	for i := range routes {
		if routes[i].Path == path {
			target := routes[i].Target
			code := routes[i].Code
			log.Printf("path: %v | target: %s | code: %v\n", path, target, code)
			http.Redirect(w, r, target, code)
		}
	}
}
