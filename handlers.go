package main

import (
	"fmt"
	"net/http"
)

func Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\"status\": \"ok\"}")
}
