package http

import (
	"encoding/json"
	"net/http"

	"github.com/MikhailKatarzhin/Level0Golang/internal/orderService"

	"github.com/gorilla/mux"
)

type OrderHandlers struct {
	orderService *orderService.OrderService
}

func NewOrderHandlers(service *orderService.OrderService) *OrderHandlers {
	return &OrderHandlers{orderService: service}
}

func (oh *OrderHandlers) GetOrderFromCacheHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderID := params["id"]

	orderData, exists := oh.orderService.GetOrderByOrderUIDFromCache(orderID)
	if !exists {
		orderData, _ = json.Marshal("Order not found in cache")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(orderData)
}

func (oh *OrderHandlers) GetOrderFromBDHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderID := params["id"]
	var orderData []byte

	order, err := oh.orderService.GetOrderByOrderUIDFromBD(orderID)
	if err != nil {
		orderData, _ = json.Marshal("Order not found in cache")
	} else {
		orderData, _ = orderService.MarshalOrder(order)
		oh.orderService.InsertOrderToCache(orderID, orderData)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(orderData)
}

func (oh *OrderHandlers) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	orderID := params["id"]

	orderData, exists := oh.orderService.GetOrderByOrderUIDFromCache(orderID)
	if !exists {
		order, err := oh.orderService.GetOrderByOrderUIDFromBD(orderID)
		if err != nil {
			orderData, _ = json.Marshal("Order not found in BD or Cache")
		} else {
			orderData, _ := orderService.MarshalOrder(order)
			oh.orderService.InsertOrderToCache(orderID, orderData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(orderData)
}
