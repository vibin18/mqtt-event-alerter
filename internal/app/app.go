package app

import (
	"fmt"
	"log/slog"
	"mqtt-event-alerter/adapters/driven/repository"
	"mqtt-event-alerter/internal/app/core"
	"time"
)

type AlertService interface {
	CreateAlert(alert core.Alert)
	GetAllAlerts() []AppData
}

type AlertApp struct {
	repos repository.Repository
}

type AppData struct {
	DateTime string
	Message  string
}

func NewAlertService(rep repository.Repository) *AlertApp {
	return &AlertApp{
		rep,
	}
}

func (a *AlertApp) CreateAlert(alert core.Alert) {
	a.repos.AddAlert(alert.DateTime, alert.Object)
}

func (a *AlertApp) GetAllAlerts() []AppData {
	alerts := a.repos.GetAllAlert()

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		slog.Error("error loading time location", slog.String("error", err.Error()))
	}
	var newData []AppData
	for _, v := range alerts {
		ndate := fmt.Sprintf("%v", v.DateTime.In(loc).Format(time.RFC1123))
		n := AppData{
			ndate,
			v.Object,
		}
		newData = append(newData, n)
	}
	return newData

}
