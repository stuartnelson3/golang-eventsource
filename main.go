package main

import (
    eventsource "github.com/antage/eventsource/http"
    "github.com/codegangsta/martini"
    "net/http"
    "strconv"
    "fmt"
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
    m.Get("/stream", es.ServeHTTP)
    m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
        card := r.FormValue("card")
        fmt.Println(card)
        es.SendMessage(card, "stream", strconv.Itoa(id))
        id++
    })
    m.Run()
}
