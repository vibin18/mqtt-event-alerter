package mqtt

import (
	"fmt"
	mq "github.com/eclipse/paho.mqtt.golang"
	"time"
)

func NewMqttConfig(listen string) *mq.ClientOptions {
	options := mq.NewClientOptions()
	options.ConnectTimeout = 30 * time.Second
	options.ConnectRetry = true
	options.AutoReconnect = true
	options.KeepAlive = 25
	options.CleanSession = true
	options.ConnectRetryInterval = 20 * time.Second
	options.PingTimeout = 30 * time.Second
	options.MaxReconnectInterval = 30 * time.Second
	options.ResumeSubs = true
	options.AddBroker(fmt.Sprintf("tcp://%v", listen))
	options.SetClientID("go_mqtt_client")

	return options

}
