package mqtt

import (
	"encoding/json"
	"fmt"
	mq "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"mqtt-event-alerter/adapters/driven/messengers"
	"mqtt-event-alerter/adapters/driven/repository"
	"mqtt-event-alerter/adapters/driven/snapshot"
	"time"
)

type MQTTHandler struct {
	Messenger  messengers.Alerter
	Repository repository.Repository
}

func NewMqttHandler(m messengers.Alerter, repo repository.Repository) MQTTHandler {
	return MQTTHandler{
		Messenger:  m,
		Repository: repo,
	}
}

func (m MQTTHandler) ConnectionHandler(client mq.Client) mq.OnConnectHandler {
	return func(client mq.Client) {
		slog.Info("connected to mqtt")
		slog.Info("Subscribing to frigate/events")
		m.Sub(client, "frigate/events")
	}
}

func (m MQTTHandler) Sub(client mq.Client, topic string) {
	token := client.Subscribe(topic, 1, m.MessagePubHandler)
	token.Wait()
	slog.Info("subscribed to topic", slog.String("topic", topic))
}

func (m MQTTHandler) MessagePubHandler(client mq.Client, msg mq.Message) {
	var events Events

	err := json.Unmarshal(msg.Payload(), &events)
	if err != nil {
		slog.Error("error unmarshalling", slog.String("error", err.Error()))
	}

	if events.Type == "new" {
		slog.Info("new event received")
		eventStartTime := events.Before.StartTime

		// TODO get location from parameter
		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			slog.Error("error loading time location", slog.String("error", err.Error()))
		}

		label := events.Before.Label
		camera := events.Before.Camera
		startTime := time.Unix(int64(eventStartTime), 0)
		contentTime := fmt.Sprintf("%v", startTime.In(loc).Format(time.RFC1123))

		slog.Info("generating snapshot")
		frigateSnapshot := snapshot.GetSnapshot("frigate.kerala.vbin.in", camera, true)
		slog.Debug("snapshot generated")

		slog.Debug("sending alert to messenger")
		m.Messenger.SendPictureAlert(label, camera, "eventid", contentTime, frigateSnapshot.Body)
		slog.Debug("adding alert entry to db")
		m.Repository.AddAlert(startTime, fmt.Sprintf("A %v detected on %v at %v", label, camera, contentTime))
	}
}

func (m MQTTHandler) ConnectionLostHandler() mq.ConnectionLostHandler {
	return func(client mq.Client, err error) {
		slog.Error("connection lost to mqtt: ", slog.String("error", err.Error()))
		timeout := 5
		ok := false
		for !ok {
			if timeout <= 0 {
				slog.Error("connection not ready after timeout, exiting..")
				return
			}
			ok = client.IsConnectionOpen()
			if !ok {
				slog.Warn("connection not ready")
				time.Sleep(1 * time.Second)
				timeout--
				slog.Info("attempting to reconnect", slog.String("attempt", string(rune(timeout))))
			}
			slog.Info("connected..")
		}
	}
}
