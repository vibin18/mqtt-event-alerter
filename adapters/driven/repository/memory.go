package repository

import (
	"mqtt-event-alerter/internal/app/core"
	"sync"
	"time"
)

var wg sync.Mutex

type MemoryRepository struct {
	data []core.Alert
}

func NewMemoryRepo() *MemoryRepository {
	return &MemoryRepository{
		data: []core.Alert{},
	}
}

func (r *MemoryRepository) AddAlert(datetime time.Time, Obj string) {
	alert := core.Alert{
		DateTime: datetime,
		Object:   Obj,
	}
	wg.Lock()
	r.data = append(r.data, alert)
	wg.Unlock()
}

func (r MemoryRepository) GetAllAlert() []core.Alert {
	wg.Lock()
	list := r.data
	wg.Unlock()
	return list
}
