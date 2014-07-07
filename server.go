package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/antage/eventsource"
	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
)

var (
	port  = flag.String("p", "8080", "the port to listen on")
	token = flag.String("token", "token123", "the app token")
	id    = 1
	es    = eventsource.New(
		&eventsource.Settings{
			Timeout:        5 * time.Second,
			CloseOnTimeout: false,
			IdleTimeout:    30 * time.Minute,
		},
		func(req *http.Request) [][]byte {
			return [][]byte{
				[]byte("X-Accel-Buffering: no"),
				[]byte("Access-Control-Allow-Origin: *"),
			}
		},
	)
)

func main() {
	flag.Parse()

	defer es.Close()

	m := pat.New()
	m.Get("/stream", tokenHandler(es.ServeHTTP))
	m.Post("/update_stream", tokenHandler(updateStream))

	handler := handlers.LoggingHandler(os.Stdout, m)

	log.Printf("listening on %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}

func tokenHandler(fn func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.FormValue("token") != *token {
			http.Error(w, "You are not authorized.", 403)
			return
		}

		fn(w, r)
	}
}

func updateStream(w http.ResponseWriter, r *http.Request) {
	es.SendEventMessage(r.FormValue("card"), r.FormValue("stream"), strconv.Itoa(id))
	id++
}
