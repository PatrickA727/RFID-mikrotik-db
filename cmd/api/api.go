package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/PatrickA727/mikrotik-db-sys/services/item"
	"github.com/gorilla/mux"
)

type APIServer struct {
	ListenAddr string
	db         *sql.DB
}

func NewAPIServer(listenAddr string, db *sql.DB) *APIServer {
	return &APIServer{
		ListenAddr: listenAddr,
		db: db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()

	subrouter_item := router.PathPrefix("/api/item").Subrouter()
	item_store := item.NewStore(s.db)
	item_handler := item.NewHandler(item_store)
	item_handler.RegisterRoutes(subrouter_item)	

	log.Println("Listening on port: ", s.ListenAddr)

	return http.ListenAndServe(s.ListenAddr, router)
}
