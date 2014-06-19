package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	eventsource "github.com/antage/eventsource"
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/cors"
)

func main() {
	id := 1
	es := eventsource.New(
		eventsource.DefaultSettings(),
		func(req *http.Request) [][]byte {
			return [][]byte{
				[]byte("X-Accel-Buffering: no"),
				[]byte("Access-Control-Allow-Origin: *"),
			}
		},
	)
	defer es.Close()

	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"http://*", "https://*"},
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Origin"},
	}))
	m.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		token := r.FormValue("token")
		if token != os.Getenv("TOKEN") {
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		log.Printf("Accepting ES connection")
		es.ServeHTTP(w, r)
	})
	m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		if token != os.Getenv("TOKEN") {
			return
		}
		card := r.FormValue("card")
		stream := r.FormValue("stream")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("foo", "bar")
		log.Printf("Sending message %s on stream %s", card, stream)
		es.SendEventMessage(card, stream, strconv.Itoa(id))
		id++
	})
	m.Run()
}
