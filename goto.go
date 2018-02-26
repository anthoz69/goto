package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var routes map[string][]route

type route struct {
	Path   string
	Target string
	Code   int
}

var port = flag.Int("port", 8080, "Port to serve goto redirect backend")

func main() {
	raw, _ := ioutil.ReadFile("redirect.yaml")
	err := yaml.Unmarshal(raw, &routes)

	log.Println(routes)

	if err != nil {
		log.Println(err)
	}

	http.HandleFunc("/", redirectHandler)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.DefaultServeMux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "could not start http server: %s\n", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "could not shutdown http server: %s\n", err)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Header.Get("X-Forwarded-Host")
	if len(host) == 0 {
		host = r.Host
	}

	path := r.URL.RequestURI()
	redirectHost := strings.TrimPrefix(host, "www.")

	routeList, ok := routes[redirectHost]
	if ok {
		for i := range routeList {
			if routeList[i].Path == path {
				target := routeList[i].Target
				code := routeList[i].Code
				log.Printf("path: %v | target: %s | code: %v\n", path, target, code)
				http.Redirect(w, r, target, code)
				return
			}
		}

	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "goto redirect backend - 404")

}

func isTLS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return false
}

func scheme(r *http.Request) string {
	if isTLS(r) {
		return "https"
	}
	return "http"
}
