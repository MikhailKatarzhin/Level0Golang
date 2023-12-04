package cache

import (
	"github.com/MikhailKatarzhin/Level0Golang/pkg/cache"
)

type Repository struct {
	CacheLRU *cache.LRUCache[string, []byte]
}

func NewOrderRepository(cache *cache.LRUCache[string, []byte]) *Repository {
	return &Repository{CacheLRU: cache}
}

func (repo *Repository) InsertOrder(orderUID string, data []byte) {
	repo.CacheLRU.Set(orderUID, data)
}

func (repo *Repository) GetOrderByUID(orderUID string) ([]byte, bool) {
	return repo.CacheLRU.Get(orderUID)
}

//func (repo *Repository) GetAllOrders() ([]model.Order, error) {
//	var orders []model.Order
//
//	return orders, nil
//}
