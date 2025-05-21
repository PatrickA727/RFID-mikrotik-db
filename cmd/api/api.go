package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/PatrickA727/mikrotik-db-sys/services/item"
	"github.com/PatrickA727/mikrotik-db-sys/services/user"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173","http://moengoet-inventory.my.id","https://app.moengoet-inventory.my.id", "http://localhost:3000","https://localhost:443","https://localhost"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "ngrok-skip-browser-warning"},
        AllowCredentials: true,  // Important for cookie authentication
    })

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

	return http.ListenAndServe(s.ListenAddr, c.Handler(router))
}
