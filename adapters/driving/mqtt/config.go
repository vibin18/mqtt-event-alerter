package mqtt

import (
	"fmt"
	mq "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"strconv"
	"time"
)

func NewMqttConfig(listen string, alive string) *mq.ClientOptions {
	ka, err := strconv.ParseInt(alive, 10, 64)
	if err != nil {
		slog.Error("error converting to int", slog.String("error", err.Error()))
	}
	options := mq.NewClientOptions()
	options.ConnectTimeout = 60 * time.Second
	options.ConnectRetry = true
	options.AutoReconnect = true
	options.KeepAlive = ka
	options.CleanSession = true
	options.ConnectRetryInterval = 20 * time.Second
	options.PingTimeout = 60 * time.Second
	options.MaxReconnectInterval = 30 * time.Second
	options.ResumeSubs = true
	options.AddBroker(fmt.Sprintf("tcp://%v", listen))
	options.SetClientID("go_mqtt_client")

	return options

}
