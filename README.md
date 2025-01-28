# MQTT EVENT ALERTER

[![license](https://img.shields.io/github/license/vibin18/mqtt-event-alerter)](https://github.com/vibin18/mqtt-event-alerter/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-vibin/mqtt--event--alerter-blue)](https://hub.docker.com/repository/docker/vibin/mqtt-event-alerter/)
![Docker Image Size](https://img.shields.io/docker/image-size/vibin/mqtt-event-alerter)

Frigate MQTT event alerter.


## Configuration

```
Usage:
  mqtt-event-alerter [OPTIONS]

Application Options:
      --discord_token=   Token of discord server [$DISCORD_TOKEN]
      --discord_channel= Discord channel id [$DISCORD_CHANNEL]
      --db_store=        Name of the sqlite db file (default: mqtt-alert.db) [$DB_NAME]
      --db_path=         Path for the sqlite db file (default: ./) [$DB_PATH]
      --listen_port=     Listening port for the web app (default: *:8090) [$LISTEN_PORT]
      --mqtt_server=     Mqtt server connection string (default: 192.168.200.75:1883) [$MQTT_SERVER]
      --log_type=        Log type, JSON or TXT (default: TXT) [$LOG_TYPE]
      --log_level=       Log level (default: INFO) [$LOG_LEVEL]

Help Options:
  -h, --help             Show this help message

```
