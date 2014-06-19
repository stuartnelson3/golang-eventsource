package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	eventsource "github.com/antage/eventsource"
	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
)

func main() {
	id := 1
	es := eventsource.New(
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
	defer es.Close()

	m := pat.New()
	m.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		token := r.FormValue("token")
		if token != os.Getenv("TOKEN") {
			return
		}
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
		log.Printf("Sending message %s on stream %s", card, stream)
		es.SendEventMessage(card, stream, strconv.Itoa(id))
		id++
	})
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("foo", "bar")
		w.Write([]byte("Hello"))
	})

	handler := handlers.LoggingHandler(os.Stdout, m)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Println("listening on 3000")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
