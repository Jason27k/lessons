package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("hello from backend 1\n"))
	})
	log.Fatal(http.ListenAndServe(":9001", nil))
}
