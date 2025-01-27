package api

import (
	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/streams", CreateStreamHandler).Methods("POST")
	router.HandleFunc("/streams/{stream_id}/start", StartStreamHandler).Methods("POST")
	router.HandleFunc("/streams/{stream_id}/stop", StopStreamHandler).Methods("POST")
	router.HandleFunc("/streams/{stream_id}", DeleteStreamHandler).Methods("DELETE")
	router.HandleFunc("/streams/{stream_id}", UpdateStreamHandler).Methods("PUT")
	router.HandleFunc("/streams", ListStreamsHandler).Methods("GET")
	router.HandleFunc("/metrics", MetricsHandler).Methods("GET")
	return router
}