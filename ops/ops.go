package ops

import (
	"encoding/json"
	"log"
)

type Params struct {
	DiscordToken   string `           long:"discord_token"     env:"DISCORD_TOKEN"  description:"Token of discord server" required:"true"`
	DiscordChannel string `           long:"discord_channel"   env:"DISCORD_CHANNEL"  description:"Discord channel id" required:"true"`
	DbStoreName    string `           long:"db_store"         env:"DB_NAME"  description:"Name of the sqlite db file" default:"mqtt-alert.db"`
	DbStorePath    string `           long:"db_path"         env:"DB_PATH"  description:"Path for the sqlite db file" default:"./"`
	ListenPort     string `           long:"listen_port"       env:"LISTEN_PORT"  description:"Listening port for the web app" default:"*:8090"`
	MQTTServer     string `           long:"mqtt_server"       env:"MQTT_SERVER"  description:"Mqtt server connection string" default:"192.168.200.75:1883"`
	MQTTKeepAlive  string `           long:"mqtt_keepalive"    env:"MQTT_KEEPALIVE"  description:"Mqtt keepalive time" default:"25"`
	LogType        string `           long:"log_type"          env:"LOG_TYPE"  description:"Log type, JSON or TXT" default:"TXT"`
	LogLevel       string `           long:"log_level"         env:"LOG_LEVEL"  description:"Log level" default:"INFO"`
}

func (o *Params) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
