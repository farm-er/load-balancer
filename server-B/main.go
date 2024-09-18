package main

import (
	"fmt"
	"log"
	"net/http"
)

func miniHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("server-B ...")
	w.Write([]byte("hello world"))
}

func main() {

	http.HandleFunc("/", miniHandler)

	fmt.Println("running on port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
