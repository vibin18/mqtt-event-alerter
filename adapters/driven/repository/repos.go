package repository

import (
	"mqtt-event-alerter/internal/app/core"
	"time"
)

type Repository interface {
	AddAlert(datetime time.Time, Obj string)
	GetAllAlert() []core.Alert
}
