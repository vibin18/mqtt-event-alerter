package api

import (
	"bytes"
	"embed"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log/slog"
	"mqtt-event-alerter/adapters/driven/messengers"
	"mqtt-event-alerter/adapters/driven/snapshot"
	"mqtt-event-alerter/internal/app"
	"mqtt-event-alerter/internal/app/core"
	"net/http"
	"time"
)

// Embed all files in the "assets" directory
//
//go:embed html/*
var htmlTemplates embed.FS

type ApiHandler struct {
	Reminders     app.AlertService
	logger        *slog.Logger
	Messenger     messengers.Alerter
	FrigateServer string
	SecureFrigate bool
}

func NewReminderWebHandler(reminders app.AlertService, logger *slog.Logger, messenger messengers.Alerter, frigate string, secure bool) *ApiHandler {
	return &ApiHandler{
		Reminders:     reminders,
		logger:        logger,
		Messenger:     messenger,
		FrigateServer: frigate,
	}
}

func (h *ApiHandler) Routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.GetHead)
	mux.Use(LoggingMiddleware(h.logger))
	mux.Get("/", h.Home)
	mux.Get("/create-alert", h.SendAlert)
	mux.Post("/send-alert-submit", h.SendAlertSubmit)
	mux.Get("/list-page", h.ListPage)
	mux.Get("/create-snapshot-alert", h.SendSnapshotAlertSubmit)

	return mux
}

func (h *ApiHandler) Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(htmlTemplates, "html/*.html")
	if err != nil {
		slog.Error("unable to parse the templates", slog.String("error", err.Error()))
	}
	buff := bytes.Buffer{}

	err = tmpl.ExecuteTemplate(&buff, "index.html", nil)
	if err != nil {
		slog.Error("Failed to execute template", slog.String("error", err.Error()))
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	_, err = buff.WriteTo(w)
	if err != nil {
		slog.Error("error writing to response", slog.String("error", err.Error()))
	}
}

func (h *ApiHandler) SendAlert(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(htmlTemplates, "html/*.html")
	if err != nil {
		slog.Error("unable to parse the templates", slog.String("error", err.Error()))
	}
	buff := bytes.Buffer{}

	err = tmpl.ExecuteTemplate(&buff, "create_alert.html", nil)
	if err != nil {
		slog.Error("Failed to execute template", slog.String("error", err.Error()))
		return
	}

	_, err = buff.WriteTo(w)
	if err != nil {
		slog.Error("error writing to response", slog.String("error", err.Error()))
	}
}

func (h *ApiHandler) SendAlertSubmit(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		slog.Error("unable to parse the form", slog.String("error", err.Error()))
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
	}
	message := r.FormValue("messageInput")
	slog.Debug("payload received to send", slog.String("message", message))
	slog.Debug("sending text message")
	h.Messenger.SendTextAlert(message)
	dateLayout := time.Now()
	h.Reminders.CreateAlert(core.Alert{DateTime: dateLayout, Object: message})
}

func (h *ApiHandler) SendSnapshotAlertSubmit(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(htmlTemplates, "html/*.html")
	if err != nil {
		slog.Error("unable to parse the templates", slog.String("error", err.Error()))
	}
	buff := bytes.Buffer{}
	msg := "manually triggered alert"
	label := "test"
	eventid := "dummy_id"
	slog.Debug("sending snapshot message")

	dateLayout := time.Now()
	formatedDate := dateLayout.Format(time.ANSIC)

	slog.Info("generating snapshot ")
	frigateSnapshot := snapshot.GetSnapshot(h.FrigateServer, "front_main_view", h.SecureFrigate)

	slog.Info(frigateSnapshot.Status)

	h.Messenger.SendPictureAlert(label, "front_main_view", eventid, formatedDate, frigateSnapshot.Body)
	h.Reminders.CreateAlert(core.Alert{DateTime: dateLayout, Object: msg})
	err = tmpl.ExecuteTemplate(&buff, "create_picture.html", nil)
	if err != nil {
		slog.Error("Failed to execute template", slog.String("error", err.Error()))
		return
	}

	_, err = buff.WriteTo(w)
	if err != nil {
		slog.Error("error writing to response", slog.String("error", err.Error()))
	}

}

func (h *ApiHandler) ListPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(htmlTemplates, "html/*.html")
	if err != nil {
		slog.Error("unable to parse the templates", slog.String("error", err.Error()))
	}
	buff := bytes.Buffer{}

	alerts := h.Reminders.GetAllAlerts()

	err = tmpl.ExecuteTemplate(&buff, "list.html", alerts)
	if err != nil {
		slog.Error("Failed to execute template", slog.String("error", err.Error()))
		return
	}

	_, err = buff.WriteTo(w)
	if err != nil {
		slog.Error("error writing to response", slog.String("error", err.Error()))
	}
}
