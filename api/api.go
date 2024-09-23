package api

import (
	"fmt"
	"net/http"
)

func getStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok")
}

func Start() {
	http.HandleFunc("/", getStatus)

	http.ListenAndServe(":8090", nil)
}
