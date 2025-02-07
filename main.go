package main

import (
	"errors"
	"fmt"
	mq "github.com/eclipse/paho.mqtt.golang"
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
	"runtime"
	"time"
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
	initArgparser()
	logger := NewLogHandler(arg.LogLevel, arg.LogType)
	slog.SetDefault(logger)
	dbPath := fmt.Sprintf("%s/%s", arg.DbStorePath, arg.DbStoreName)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqliteRepo := repository.NewSqlRepo(db)
	//repo := repository.NewMemoryRepo()

	discord := messengers.NewDiscordMessenger(arg.DiscordToken, arg.DiscordChannel)
	service := app.NewAlertService(sqliteRepo)
	handler := api.NewReminderWebHandler(service, logger, discord)
	mqttHandler := mqtt.NewMqttHandler(discord, sqliteRepo)
	mqttClientOptions := mqtt.NewMqttConfig(arg.MQTTServer, arg.MQTTKeepAlive)

	slog.Info("creating new mqtt client")
	client := mq.NewClient(mqttClientOptions)
	mqttClientOptions.OnConnect = mqttHandler.ConnectionHandler(client)
	mqttClientOptions.OnConnectionLost = mqttHandler.ConnectionLostHandler()
	slog.Info("connecting to mqtt server")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	go func() {
		timeout := 1000

		ok := false
		for !ok {
			if timeout <= 0 {
				slog.Info("connection not ready after timeout, exiting..")
				return
			}
			ok = client.IsConnectionOpen()
			if !ok {
				slog.Info("mqtt connection not ready")
				slog.Info("retrying connection..")
				time.Sleep(10 * time.Second)
				timeout--
			}
			slog.Info("connected to mqtt")

		}

		mqttHandler.Sub(client, "frigate/events")
	}()
	slog.Info("Starting Golang Application on port *:8090", slog.String("arch", runtime.GOARCH), slog.String("compiler", runtime.Compiler), slog.String("version", runtime.Version()))
	err = http.ListenAndServe(":8090", handler.Routes())
	if err != nil {
		log.Panic(err)
	}
}
