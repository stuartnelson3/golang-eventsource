package main

import (
    eventsource "github.com/antage/eventsource/http"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini"
    "net/http"
    "strconv"
    // "os"
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
        AllowOrigins:     []string{"http://*", "https://*"},
        AllowMethods:     []string{"GET"},
        AllowHeaders:     []string{"Origin"},
    }))
    m.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
        // token := r.FormValue("token")
        // if token != os.Getenv("TOKEN") {
        //     return
        // }
        es.ServeHTTP(w, r)
    })
    m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
        // token := r.FormValue("token")
        // if token != os.Getenv("TOKEN") {
        //     return
        // }
        card := r.FormValue("card")
        stream := r.FormValue("stream")
        es.SendMessage(card, stream, strconv.Itoa(id))
        id++
    })
    m.Run()
}
