package api

import (
	"fmt"
	"net/http"
)

func get_status(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok")
}

func Start() {
	http.HandleFunc("/", get_status)

	http.ListenAndServe(":8090", nil)
}
