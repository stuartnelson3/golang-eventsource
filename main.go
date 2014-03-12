package main

import (
    eventsource "github.com/antage/eventsource/http"
    "github.com/codegangsta/martini"
    "net/http"
    "strconv"
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
    // TODO: add auth and all that jazz
    m.Get("/stream", es.ServeHTTP)
    m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
        card := r.FormValue("card")
        stream := r.FormValue("stream")
        es.SendMessage(card, stream, strconv.Itoa(id))
        id++
    })
    m.Run()
}
