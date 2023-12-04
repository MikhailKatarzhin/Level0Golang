package http

import (
	"fmt"
	"net/http"

	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService"
	"github.com/MikhailKatarzhin/Level0Golang/pkg/logger"

	"github.com/gorilla/mux"
)

type Server struct {
	router    *mux.Router
	orderServ *orderService.OrderService
}

func NewServer(orderService *orderService.OrderService) *Server {
	server := &Server{
		router:    mux.NewRouter(),
		orderServ: orderService,
	}
	server.routes()
	return server
}

func (s *Server) routes() {
	orderHandlers := NewOrderHandlers(s.orderServ)

	s.router.HandleFunc("/orders/get", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./././web/html/orderById.html")
	})
	s.router.HandleFunc("/orders/getCache", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./././web/html/orderByIdFromCache.html")
	})
	s.router.HandleFunc("/orders/getBD", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./././web/html/orderByIdFromBD.html")
	})

	s.router.HandleFunc("/orders/get/{id}", orderHandlers.GetOrderHandler).Methods("GET")
	s.router.HandleFunc("/orders/getCache/{id}", orderHandlers.GetOrderFromCacheHandler).Methods("GET")
	s.router.HandleFunc("/orders/getBD/{id}", orderHandlers.GetOrderFromBDHandler).Methods("GET")
}

func (s *Server) Start(port string) {
	logger.L().Info(fmt.Sprintf("Starting server on port %s", port))
	err := http.ListenAndServe(port, s.router)
	if err != nil {
		logger.L().Fatal(fmt.Sprintf("Error starting server: %s", err))
	}
}
