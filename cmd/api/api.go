package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/PatrickA727/mikrotik-db-sys/services/item"
	"github.com/PatrickA727/mikrotik-db-sys/services/user"
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

	// Init stores
	item_store := item.NewStore(s.db)
	user_store := user.NewStore(s.db)

	subrouter_item := router.PathPrefix("/api/item").Subrouter()
	item_handler := item.NewHandler(item_store, user_store)
	item_handler.RegisterRoutes(subrouter_item)	

	subrouter_user := router.PathPrefix("/api/user").Subrouter()
	user_handler := user.NewHandler(user_store)
	user_handler.RegisterRoutes(subrouter_user)

	log.Println("Listening on port: ", s.ListenAddr)

	return http.ListenAndServe(s.ListenAddr, router)
}
