package main

import (
        "fmt"
        "log"
        "net/http"
        "os"
)

const ver = "ver1.1"

func main() {
        http.HandleFunc("/", HelloServer)
        fmt.Printf("Starting server at port 80\n")
        if err := http.ListenAndServe(":80", nil); err != nil {
                log.Fatal(err)
        }
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
        name, err := os.Hostname()
        if err != nil {
                panic(err)
        }
        fmt.Fprint(w, "Hello world from " + name + " " + ver + "\n")
}