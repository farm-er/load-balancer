package main

import (
	"fmt"
	"log"
	"net/http"
)

func miniHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("server-A ...")
	w.Write([]byte("hello world"))
}

func main() {

	http.HandleFunc("/", miniHandler)

	fmt.Println("running on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
