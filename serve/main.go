package main

import (
	"fmt"
	"net/http"
)

func main() {
	listenAndServe(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "hello world\n")
	}))
}
