package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"log/slog"
	"mqtt-event-alerter/adapters/driven/messengers"
	"mqtt-event-alerter/adapters/driven/repository"
	"mqtt-event-alerter/adapters/driving/api"
	"mqtt-event-alerter/adapters/driving/mqtt"
	"mqtt-event-alerter/internal/app"
	"mqtt-event-alerter/ops"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	topic    = "frigate/events"
	clientID = "mqtt-subscriber"
)

var (
	argparser *flags.Parser
	arg       ops.Params
)

func initArgparser() {
	argparser = flags.NewParser(&arg, flags.Default)
	_, err := argparser.Parse()

	// check if there is a parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

func main() {
	// Create a cancelable context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Init Argument parser
	initArgparser()

	// Init logger
	logger := NewLogHandler(arg.LogLevel, arg.LogType)
	slog.SetDefault(logger)

	// SQLITE file
	dbPath := fmt.Sprintf("%s/%s", arg.DbStorePath, arg.DbStoreName)

	// Init DB
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Create Repository instance
	sqliteRepo := repository.NewSqlRepo(db)
	//repo := repository.NewMemoryRepo()
	// Create Messenger instance
	discord := messengers.NewDiscordMessenger(arg.DiscordToken, arg.DiscordChannel)
	// Initialize alert service
	service := app.NewAlertService(sqliteRepo)

	// Initialize API
	handler := api.NewReminderWebHandler(service, logger, discord, arg.FrigateServer, arg.SecureFrigate)

	// Create MQTT client
	mqttClient := mqtt.NewMQTTClient(arg.MQTTServer, topic, clientID, nil, discord, sqliteRepo, arg.FrigateServer, arg.SecureFrigate)

	// Initialize MQTT
	go mqttClient.Run()

	// Initialize API
	var server = &http.Server{
		Addr:    arg.ListenPort,
		Handler: handler.Routes(),
	}

	go func() {

		slog.Info("Starting Golang Application on port *:8090", slog.String("arch", runtime.GOARCH), slog.String("compiler", runtime.Compiler), slog.String("version", runtime.Version()))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panic(err)
		}
	}()

	// Capture termination signals (SIGINT, SIGTERM, SIGTSTP)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGTSTP)

	// Wait for a termination signal
	sig := <-sigChan
	slog.Info("shutting down, ", slog.String("received signal", fmt.Sprint(sig)))

	// Cancel context to notify all goroutines
	cancel()

	// Gracefully shut down MQTT client
	mqttClient.Stop()

	if err := server.Shutdown(ctx); err != nil {
		slog.Info("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	// Send SIGTERM to the process
	slog.Info("sending SIGTERM signal...")
	p, err := os.FindProcess(os.Getpid()) // Get current process
	if err != nil {
		slog.Info("error finding process:", slog.String("error", err.Error()))
		return
	}
	p.Signal(syscall.SIGTERM)
	slog.Info("Cleanup complete. Exiting.")

}
