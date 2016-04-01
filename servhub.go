package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"./app"
)

func main() {
	s := app.NewServer()
	r := mux.NewRouter()
	r.HandleFunc("/api/v1.0/service", s.CreateServHandler).Methods("POST")
	r.HandleFunc("/api/v1.0/service", s.UpdateServHandler).Methods("PUT")
	r.HandleFunc("/api/v1.0/service", s.DeleteServHandler).Methods("DELETE")
	r.HandleFunc("/api/v1.0/service/config", s.SetConfigHandler).Methods("POST")
	r.HandleFunc("/api/v1.0/service/config", s.RmvConfigHandler).Methods("DELETE")
	http.Handle("/", r)
	http.ListenAndServe(":7171", nil)
}