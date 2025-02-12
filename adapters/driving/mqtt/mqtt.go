package mqtt

import (
	"encoding/json"
	"fmt"
	mq "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"mqtt-event-alerter/adapters/driven/messengers"
	"mqtt-event-alerter/adapters/driven/repository"
	"mqtt-event-alerter/adapters/driven/snapshot"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	retryDelay   = 5 * time.Second  // Delay before retrying connection
	maxRetries   = 100              // Max retries before giving up
	reconnectGap = 10 * time.Second // Wait time before trying to reconnect
)

// MQTTClient wraps the MQTT client
type MQTTClient struct {
	Messenger      messengers.Alerter
	Repository     repository.Repository
	FrigateServer  string
	SecureFrigate  bool
	client         mq.Client
	broker         string
	topic          string
	clientID       string
	messageHandler mq.MessageHandler
	stopChan       chan struct{}
}

// NewMQTTClient creates a new MQTTClient instance
func NewMQTTClient(broker, topic, clientID string, handler mq.MessageHandler, m messengers.Alerter, repo repository.Repository, frigate string, secure bool) *MQTTClient {
	return &MQTTClient{
		broker:         broker,
		topic:          topic,
		clientID:       clientID,
		messageHandler: handler,
		Messenger:      m,
		Repository:     repo,
		FrigateServer:  frigate,
		SecureFrigate:  secure,
		stopChan:       make(chan struct{}),
	}
}

// ConnectWithRetry connects to the MQTT broker with retry logic
func (m *MQTTClient) ConnectWithRetry() {
	opts := mq.NewClientOptions().
		AddBroker(m.broker).
		SetClientID(m.clientID).
		SetDefaultPublishHandler(m.messageHandler).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(reconnectGap)

	for i := 1; i <= maxRetries; i++ {
		m.client = mq.NewClient(opts)
		token := m.client.Connect()
		if token.Wait() && token.Error() == nil {
			slog.Info("connected to MQTT broker")
			return
		}

		slog.Info("MQTT connection failed ", slog.String(fmt.Sprint(i), fmt.Sprint(maxRetries)), slog.String("token error", token.Error().Error()))
		time.Sleep(retryDelay)
	}

	slog.Error("Max retries reached, could not connect to MQTT broker")
	os.Exit(1)
}

// Subscribe subscribes to the configured topic
func (m *MQTTClient) Subscribe() {
	token := m.client.Subscribe(m.topic, 1, m.MessageHandler)
	if token.Wait() && token.Error() != nil {
		slog.Info("Subscription failed", slog.String("topic", token.Error().Error()))
		os.Exit(1)
	}
	slog.Info("Subscribed to", slog.String("topic", m.topic))
}

// Run starts the MQTT client and listens for termination signals
func (m *MQTTClient) Run() {
	slog.Info("starting mqtt handler")
	m.ConnectWithRetry()
	slog.Info("subscribing to mqtt topic")
	m.Subscribe()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down MQTT client...")
	m.client.Disconnect(250)
	slog.Info("MQTT client disconnected. Exiting.")
}

func (m *MQTTClient) MessageHandler(client mq.Client, msg mq.Message) {
	var events Events

	err := json.Unmarshal(msg.Payload(), &events)
	if err != nil {
		slog.Error("error unmarshalling", slog.String("error", err.Error()))
	}

	if events.Type == "new" {
		slog.Info("new mqtt event received")
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

		frigateSnapshot := snapshot.GetSnapshot(m.FrigateServer, camera, m.SecureFrigate)

		slog.Info("sending alert to messenger")
		m.Messenger.SendPictureAlert(label, camera, "eventid", contentTime, frigateSnapshot.Body)
		slog.Info("adding alert entry to database")
		m.Repository.AddAlert(startTime, fmt.Sprintf("A %v detected on %v at %v", label, camera, contentTime))
	}
}

func (m *MQTTClient) Stop() {
	close(m.stopChan)
}
