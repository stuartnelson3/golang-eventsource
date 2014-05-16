package main

import (
	"fmt"
	eventsource "github.com/antage/eventsource"
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/cors"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	id := 1

	esMap := make(map[string]eventsource.EventSource)

	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins: []string{"*"},
	}))

	// Monitor and remove any dead es
	go MonitorAndReap(esMap)

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("token") != os.Getenv("TOKEN") {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	})

	m.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		if token != os.Getenv("TOKEN") {
			return
		}

		stream := r.FormValue("stream")
		if stream == "" {
			return
		}
		// Loop through all stream identifiers
		// If es already exists, add connection to it
		if es, prs := esMap[stream]; prs {
			es.ServeHTTP(w, r)
			fmt.Println("Using existing stream: ", stream)
		} else {
			// If es does not exist, create and add connection to it
			es := eventsource.New(
				&eventsource.Settings{
					IdleTimeout:    30 * time.Minute,
					Timeout:        2 * time.Second,
					CloseOnTimeout: true,
				},
				func(req *http.Request) [][]byte {
					return [][]byte{
						[]byte("X-Accel-Buffering: no"),
						[]byte("Access-Control-Allow-Origin: *"),
					}
				},
			)
			esMap[stream] = es
			fmt.Println("Created new stream: ", stream)
			es.ServeHTTP(w, r)
			es.SendRetryMessage(100 * time.Millisecond)
		}
	})

	m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")
		if token != os.Getenv("TOKEN") {
			return
		}
		card := r.FormValue("card")
		stream := r.FormValue("stream")
		esMap[stream].SendEventMessage(card, stream, strconv.Itoa(id))
		id++
	})
	m.Run()
}

func MonitorAndReap(esMap map[string]eventsource.EventSource) {
	for {
		for stream, es := range esMap {
			if es.ConsumersCount() == 0 {
				es.Close()
				delete(esMap, stream)
				fmt.Println("Removed: ", stream)
			}
		}
		time.Sleep(time.Minute * 5)
	}
}
