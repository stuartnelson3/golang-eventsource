package main

import (
    eventsource "github.com/stuartnelson3/eventsource/http"
    "github.com/martini-contrib/cors"
    "github.com/codegangsta/martini"
    "net/http"
    "strconv"
    "time"
    "fmt"
    "os"
)

func main() {
    id := 1

    esMap := make(map[string]eventsource.EventSource)

    m := martini.Classic()
    m.Use(cors.Allow(&cors.Options{
        AllowOrigins:     []string{"http://*", "https://*"},
        AllowMethods:     []string{"GET"},
        AllowHeaders:     []string{"Origin"},
    }))

    // Monitor and remove any dead es
    go MonitorAndReap(esMap)

    m.Get("/stream", func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        token := r.FormValue("token")
        if token != os.Getenv("TOKEN") {
            return
        }

        streams := r.Form["streams[]"]
        // Loop through all stream identifiers
        for i := 0; i < len(streams); i++ {
            stream := streams[i]
            // If es already exists, add connection to it
            if es, prs := esMap[stream]; prs {
                es.ServeHTTP(w, r)
                fmt.Println("Using existing stream: ", stream)
            } else {
                // If es does not exist, create and add connection to it
                es := eventsource.New(
                    eventsource.DefaultSettings(),
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
            }
        }
    })

    m.Post("/update_stream", func(w http.ResponseWriter, r *http.Request) {
        token := r.FormValue("token")
        if token != os.Getenv("TOKEN") {
            return
        }
        card := r.FormValue("card")
        stream := r.FormValue("stream")
        esMap[stream].SendMessage(card, stream, strconv.Itoa(id))
        id++
    })
    m.Run()
}

func MonitorAndReap(esMap map[string]eventsource.EventSource) {
    for {
        for stream, es := range esMap {
            // This is never returning 0, so it's never removing users
            // I don't think the user is ever signalling that it is stalled
            if es.ConsumersCount() == 0 {
                es.Close()
                delete(esMap, stream)
                fmt.Println("Removed: ", stream)
            }
        }
        time.Sleep(time.Second)
        // time.Sleep(time.Minute)
    }
}
