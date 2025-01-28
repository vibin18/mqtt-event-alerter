package repository

import (
	"gorm.io/gorm"
	"log/slog"
	"mqtt-event-alerter/internal/app/core"
	"time"
)

type AlertData struct {
	gorm.Model
	TimeDate int64
	Message  string
}

type SqlRepository struct {
	db *gorm.DB
}

func NewSqlRepo(db *gorm.DB) *SqlRepository {
	err := db.AutoMigrate(&AlertData{})
	if err != nil {
		slog.Error("error migrating db", slog.String("error", err.Error()))
	}

	return &SqlRepository{
		db,
	}
}

func (r *SqlRepository) AddAlert(datetime time.Time, Obj string) {

	unixTime := datetime.Unix()

	inputData := AlertData{
		TimeDate: unixTime,
		Message:  Obj,
	}
	r.db.Create(&inputData)
}

func (r *SqlRepository) GetAllAlert() []core.Alert {
	var alerts []AlertData
	var coreAlerts []core.Alert

	err := r.db.Order("time_date DESC").Find(&alerts).Error
	if err != nil {
		slog.Error("error getting all alerts from db", slog.String("error", err.Error()))
	}

	for _, alert := range alerts {
		coreAlert := core.Alert{
			DateTime: time.Unix(alert.TimeDate, 0),
			Object:   alert.Message,
		}

		coreAlerts = append(coreAlerts, coreAlert)
	}
	return coreAlerts

}
